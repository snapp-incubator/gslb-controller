/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
)

var _ = Describe("Gslb controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		GslbName      = "test-gslb"
		GslbNamespace = "default"
		timeout       = time.Second * 10
		duration      = time.Second * 10
		interval      = time.Millisecond * 250
	)

	gslbMeta := &gslbv1alpha1.Gslb{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "gslb.snappcloud.io/v1alpha1",
			Kind:       "Gslb",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      GslbName,
			Namespace: GslbNamespace,
		},
	}

	gslbContent0TypeMeta := metav1.TypeMeta{
		APIVersion: "gslb.snappcloud.io/v1alpha1",
		Kind:       "GslbContent",
	}

	var gslbContent0 *gslbv1alpha1.GslbContent

	Context("When createing a Gslb", func() {

		It("Should create a Gslbcontent for each Gslb backend", func() {
			By("By creating a full Gslb", func() {
				ctx := context.Background()
				gslb := &gslbv1alpha1.Gslb{
					TypeMeta:   gslbMeta.TypeMeta,
					ObjectMeta: gslbMeta.ObjectMeta,
					Spec: gslbv1alpha1.GslbSpec{
						ServiceName: "integration-test",
						Backends: []gslbv1alpha1.Backend{
							{
								Name:   "complete",
								Host:   "google.com",
								Weight: "1",
								Probe: gslbv1alpha1.Probe{
									Handler: gslbv1alpha1.Handler{
										HTTPGet: &gslbv1alpha1.HTTPGetAction{
											Scheme: "http",
											Path:   "/",
											Port:   80,
											Host:   "google.com",
										},
									},
									TimeoutSeconds: 3,
									PeriodSeconds:  3,
								},
							},
							// {
							// 	Name:   "minimum",
							// 	Host:   "bing.com",
							// 	Probe: gslbv1alpha1.Probe{
							// 		Handler: gslbv1alpha1.Handler{
							// 			HTTPGet: &gslbv1alpha1.HTTPGetAction{
							// 				Scheme: "http",
							// 			},
							// 		},
							// 	},
							// },
						},
					},
				}
				Expect(k8sClient.Create(ctx, gslb)).Should(Succeed())

				gslbLookupKey := types.NamespacedName{Name: GslbName, Namespace: GslbNamespace}
				createdGslb := &gslbv1alpha1.Gslb{}
				Expect(k8sClient.Get(ctx, gslbLookupKey, createdGslb)).Should(Succeed())

				// Let's make sure created object, matches desired one.
				Expect(createdGslb.TypeMeta).Should(Equal(gslb.TypeMeta))
				Expect(createdGslb.GetObjectMeta()).Should(Equal(gslb.GetObjectMeta()))
				Expect(createdGslb.Spec).Should(Equal(gslb.Spec))

				expectedGslbContent0 := &gslbv1alpha1.GslbContent{
					TypeMeta: gslbContent0TypeMeta,
					ObjectMeta: metav1.ObjectMeta{
						Name:   "gslb" + "-" + string(gslb.Spec.ServiceName) + "-" + string(createdGslb.GetUID()) + "-" + gslb.Spec.GetBackends()[0].Name,
						Labels: labelsForGslbcon(gslb.GetName(), gslb.GetNamespace()),
					},
					Spec: gslbv1alpha1.GslbContentSpec{
						ServiceName: gslb.Spec.ServiceName,
						Backend:     gslb.Spec.GetBackends()[0],
					},
				}
				createdGslbContent0 := &gslbv1alpha1.GslbContent{}
				// time.Sleep(5 * time.Second)
				gslbContentLookupKey := types.NamespacedName{Name: expectedGslbContent0.GetName(), Namespace: ""}
				Eventually(func() bool {
					err := k8sClient.Get(ctx, gslbContentLookupKey, createdGslbContent0)
					return err == nil
				}, timeout, interval).Should(BeTrue())
				// TODO: Why the follwoing doesn't work!
				// Eventually(k8sClient.Get(ctx, gslbContentLookupKey, createdGslbContent0), timeout, interval).Should(BeNil())
				// TODO: Why also Succeed() doesn't work?!

				// Expect(createdGslbContent0.TypeMeta).Should(Equal(expectedGslbContent0.TypeMeta))
				Expect(createdGslbContent0.GetName()).Should(Equal(expectedGslbContent0.GetName()))
				Expect(createdGslbContent0.GetLabels()).Should(Equal(expectedGslbContent0.GetLabels()))
				Expect(createdGslbContent0.Spec).Should(Equal(expectedGslbContent0.Spec))

				// By("By checking the Gslb backend weight")
				// Consistently(func() (string, error) {
				// 	err := k8sClient.Get(ctx, gslbLookupKey, createdGslb)
				// 	if err != nil {
				// 		return "", err
				// 	}
				// 	return createdGslb.Spec.GetBackends()[0].Weight, nil
				// }, duration, interval).Should(Equal("1"))
			})
		})
	})

	Context("When updating a Gslb", func() {
		It("Should update corresponding GslbContent for each Gslb backend", func() {
			ctx := context.Background()
			gslbLookupKey := types.NamespacedName{Name: GslbName, Namespace: GslbNamespace}
			createdGslb := &gslbv1alpha1.Gslb{}
			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			Expect(k8sClient.Get(ctx, gslbLookupKey, createdGslb)).Should(Succeed())
			updatedGslb := createdGslb.DeepCopy()
			updatedGslb.Spec.ServiceName = "bing"
			Expect(k8sClient.Update(ctx, updatedGslb)).Should(Succeed())

			By("Deleting the old GslbContent")
			oldGslbContentName := "gslbs" + "-" + string(createdGslb.Spec.ServiceName) + "-" + string(createdGslb.GetUID()) + "-" + createdGslb.Spec.GetBackends()[0].Name
			gslbContentLookupKey := types.NamespacedName{Name: oldGslbContentName, Namespace: ""}
			createdGslbContent0 := &gslbv1alpha1.GslbContent{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, gslbContentLookupKey, createdGslbContent0)
				return errors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Creating a new GslbContent")
			newGslbContentName := "gslb" + "-" + string(updatedGslb.Spec.ServiceName) + "-" + string(updatedGslb.GetUID()) + "-" + updatedGslb.Spec.GetBackends()[0].Name
			gslbContentLookupKey = types.NamespacedName{Name: newGslbContentName, Namespace: ""}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, gslbContentLookupKey, createdGslbContent0)
				return err == nil
			}, timeout, interval).Should(BeTrue())
			expectedGslbContent0 := &gslbv1alpha1.GslbContent{
				TypeMeta: gslbContent0TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:   newGslbContentName,
					Labels: labelsForGslbcon(updatedGslb.GetName(), updatedGslb.GetNamespace()),
				},
				Spec: gslbv1alpha1.GslbContentSpec{
					ServiceName: updatedGslb.Spec.ServiceName,
					Backend:     updatedGslb.Spec.GetBackends()[0],
				},
			}

			// Expect(createdGslbContent0.TypeMeta).Should(Equal(expectedGslbContent0.TypeMeta))
			Expect(createdGslbContent0.GetName()).Should(Equal(expectedGslbContent0.GetName()))
			Expect(createdGslbContent0.GetLabels()).Should(Equal(expectedGslbContent0.GetLabels()))
			Expect(createdGslbContent0.Spec).Should(Equal(expectedGslbContent0.Spec))
			gslbContent0 = createdGslbContent0
		})
	})

	Context("When deleting a managed GslbContent", func() {
		It("Should recreate the GslbContent", func() {
			ctx := context.Background()
			// By("By deleting a new Gslb")
			gslbContent0Meta := &gslbv1alpha1.GslbContent{
				TypeMeta: gslbContent0TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name: gslbContent0.GetName(),
				},
			}

			Expect(k8sClient.Delete(ctx, gslbContent0Meta)).To(Succeed())

			createdGslbContent0 := &gslbv1alpha1.GslbContent{}
			gslbContentLookupKey := types.NamespacedName{Name: gslbContent0.GetName(), Namespace: ""}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, gslbContentLookupKey, createdGslbContent0)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			expectedGslbContent0 := &gslbv1alpha1.GslbContent{
				TypeMeta: gslbContent0TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:   gslbContent0.GetName(),
					Labels: gslbContent0.GetLabels(),
				},
				Spec: gslbContent0.Spec,
			}

			// Expect(createdGslbContent0.TypeMeta).Should(Equal(expectedGslbContent0.TypeMeta))
			Expect(createdGslbContent0.GetName()).Should(Equal(expectedGslbContent0.GetName()))
			Expect(createdGslbContent0.GetLabels()).Should(Equal(expectedGslbContent0.GetLabels()))
			Expect(createdGslbContent0.Spec).Should(Equal(expectedGslbContent0.Spec))
			gslbContent0 = createdGslbContent0
		})
	})

	Context("When updating a managed GslbContent", func() {
		It("Should revert the changes on the GslbContent", func() {
			ctx := context.Background()
			updatedGslbContent0 := gslbContent0.DeepCopy()

			updatedGslbContent0.Spec.ServiceName = "ali"
			Expect(k8sClient.Update(ctx, updatedGslbContent0)).Should(Succeed())

			createdGslbContent0 := &gslbv1alpha1.GslbContent{}
			gslbContentLookupKey := types.NamespacedName{Name: gslbContent0.GetName(), Namespace: ""}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, gslbContentLookupKey, createdGslbContent0)
				if err != nil {
					return false
				}
				// Wait until reverted
				if createdGslbContent0.Spec.ServiceName != "bing" {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			expectedGslbContent0 := &gslbv1alpha1.GslbContent{
				TypeMeta: gslbContent0TypeMeta,
				ObjectMeta: metav1.ObjectMeta{
					Name:   gslbContent0.GetName(),
					Labels: gslbContent0.GetLabels(),
				},
				Spec: gslbContent0.Spec,
			}

			// 		Expect(createdGslbContent0.TypeMeta).Should(Equal(expectedGslbContent0.TypeMeta))
			Expect(createdGslbContent0.GetName()).Should(Equal(expectedGslbContent0.GetName()))
			Expect(createdGslbContent0.GetLabels()).Should(Equal(expectedGslbContent0.GetLabels()))
			Expect(createdGslbContent0.Spec).Should(Equal(expectedGslbContent0.Spec))

		})
	})

	Context("When deleting a Gslb", func() {
		It("Should delete corresponding GslbContent for each Gslb backend", func() {
			By("By deleting a Gslb", func() {
				ctx := context.Background()
				Expect(k8sClient.Delete(ctx, gslbMeta)).Should(Succeed())
				Eventually(func() error {
					gslbContentList := &gslbv1alpha1.GslbContentList{}
					listOpts := []client.ListOption{
						client.MatchingLabels(labelsForGslbcon(GslbName, GslbNamespace)),
					}

					if err := k8sClient.List(ctx, gslbContentList, listOpts...); err != nil {
						return err
					}
					if len(gslbContentList.Items) != 0 {
						return fmt.Errorf("Remaining orphan Gslbcontents: %v", gslbContentList.Items)
					}
					return nil
				}, timeout, interval).Should(BeNil())
			})
		})
	})

})
