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

package v1alpha1

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("gslb webhook", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		fooGslbName      = "foo-gslb"
		fooGslbNameSpace = "default"
		barGslbName      = "bar-gslb"
		barGslbNameSpace = "default"
		bazGslbName      = "baz-gslb"
		bazGslbNameSpace = "default"
		timeout          = time.Second * 10
		duration         = time.Second * 10
		interval         = time.Millisecond * 250
	)
	var (
		err  error
		gslb *Gslb
		ctx  = context.Background()
	)

	gslbTypeMeta := metav1.TypeMeta{
		APIVersion: "gslb.snappcloud.io/v1alpha1",
		Kind:       "Gslb",
	}

	fooGslbMeta := &Gslb{
		TypeMeta: gslbTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      fooGslbName,
			Namespace: fooGslbNameSpace,
		},
	}

	barGslbMeta := &Gslb{
		TypeMeta: gslbTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      barGslbName,
			Namespace: barGslbNameSpace,
		},
	}

	bazGslbMeta := &Gslb{
		TypeMeta: gslbTypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:      bazGslbName,
			Namespace: bazGslbNameSpace,
		},
	}

	AfterEach(func() {
		err = k8sClient.Delete(ctx, fooGslbMeta)
		if err != nil {
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		}
		err = k8sClient.Delete(ctx, barGslbMeta)
		if err != nil {
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		}
		err = k8sClient.Delete(ctx, bazGslbMeta)
		if err != nil {
			Expect(errors.IsNotFound(err)).Should(BeTrue())
		}
	})

	Context("When createing a Gslb", func() {

		It("Should reject if there is a repatative backend name", func() {
			By("Creating a Gslb with two repatative backend names")
			fooGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test",
					Backends: []Backend{
						{
							Name:   "complete",
							Host:   "google.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
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
						{
							Name:   "complete",
							Host:   "bing.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "bing.com",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, fooGslb)
			Expect(string(errors.ReasonForError(err))).Should(Equal("duplicate backend name found: complete. All backend names must be unique"))

		})

		It("Should reject if serviceName exits in claimedServiceNames", func() {
			By("first creating fooGslb service")
			gslb = &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test",
					Backends: []Backend{
						{
							Name:   "google",
							Host:   "google.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
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
					},
				},
			}
			err = k8sClient.Create(ctx, gslb)
			Expect(err).To(BeNil())

			By("then creating  barGslb with existing serviceName")
			gslb = &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: barGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test",
					Backends: []Backend{
						{
							Name:   "bing",
							Host:   "bing.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "bing.com",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, gslb)
			Expect(string(errors.ReasonForError(err))).Should(Equal("'integration-test' serviceName is already claimed. please try another serviceName"))
		})

	})

	Context("When deleting a Gslb", func() {

		It("Should remove the service name from claimedServiceNames", func() {
			By("first creating fooGslb service")
			gslb = &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test",
					Backends: []Backend{
						{
							Name:   "google",
							Host:   "google.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
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
					},
				},
			}
			Expect(k8sClient.Create(ctx, gslb)).To(Succeed())

			By("Deleting the created Gslb")
			Expect(k8sClient.Delete(ctx, fooGslbMeta)).Should(Succeed())

			By("letting barGslb with the same serviceName to be created")
			gslb = &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: barGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, gslb)
			Expect(err).To(BeNil())

		})
	})

	Context("When Updating a Gslb via renaming a serviceName", func() {
		It("should reject it if it exists in claimedServiceNames", func() {
			By("first, creating two Gslbs with two different serviceNames")
			fooGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test-1",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			barGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: barGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test-2",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, fooGslb)
			Expect(err).To(BeNil())
			err = k8sClient.Create(ctx, barGslb)
			Expect(err).To(BeNil())

			By("Renaming the barGslb ServiceName to fooGslb ServiceName")
			gslbLookupKey := types.NamespacedName{Name: barGslb.GetName(), Namespace: barGslb.GetNamespace()}
			err = k8sClient.Get(ctx, gslbLookupKey, barGslb)
			Expect(err).To(BeNil())

			barGslb.Spec.ServiceName = "integration-test-1"

			err = k8sClient.Update(ctx, barGslb)
			Expect(string(errors.ReasonForError(err))).Should(Equal("'integration-test-1' serviceName is already claimed. please try another serviceName"))
		})

		It("should add new, and release old service name from claimedServiceNames", func() {
			By("first, creating fooGslb and barGslb with two different serviceNames")
			fooGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test-1",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			barGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: barGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test-2",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, fooGslb)
			Expect(err).To(BeNil())
			err = k8sClient.Create(ctx, barGslb)
			Expect(err).To(BeNil())

			By("editing fooGslb serviceName to a new ServiceName")
			gslbLookupKey := types.NamespacedName{Name: fooGslbName, Namespace: fooGslbNameSpace}
			gslb = &Gslb{}
			err = k8sClient.Get(ctx, gslbLookupKey, gslb)
			Expect(err).To(BeNil())

			gslb.Spec.ServiceName = "integration-test-3"
			err = k8sClient.Update(ctx, gslb)
			Expect(err).To(BeNil())

			By("creating bazGslb with the old fooGslb servieName")
			bazGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: bazGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test-1",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, bazGslb)
			Expect(err).To(BeNil())

		})
	})

	Context("When Updating Gslb via updating backends", func() {

		It("Should reject if there is a repatative backend name when adding a new backend", func() {
			By("Creating fooGslb")
			fooGslb := &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test-1",
					Backends: []Backend{
						{
							Name:   "mci",
							Host:   "mci.ir",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "mci.ir",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, fooGslb)
			Expect(err).To(BeNil())

			gslbLookupKey := types.NamespacedName{Name: fooGslbName, Namespace: fooGslbNameSpace}
			err = k8sClient.Get(ctx, gslbLookupKey, gslb)
			Expect(err).To(BeNil())

			gslb.Spec.Backends = []Backend{
				gslb.Spec.GetBackends()[0],
				{
					Name:   gslb.Spec.GetBackends()[0].Name,
					Host:   "mci.ir",
					Weight: "1",
					Probe: Probe{
						Handler: Handler{
							HTTPGet: &HTTPGetAction{
								Scheme: "http",
								Path:   "/",
								Port:   80,
								Host:   "mci.ir",
							},
						},
						TimeoutSeconds: 3,
						PeriodSeconds:  3,
					},
				},
			}
			err = k8sClient.Update(ctx, gslb)
			Expect(string(errors.ReasonForError(err))).Should(Equal("duplicate backend name found: " + gslb.Spec.GetBackends()[0].Name + ". All backend names must be unique"))
		})

		It("Should reject if there is a repatative backend name when renaming an existing backend", func() {
			By("creating a Gslb with multiple backends")

			gslb = &Gslb{
				TypeMeta:   gslbTypeMeta,
				ObjectMeta: fooGslbMeta.ObjectMeta,
				Spec: GslbSpec{
					ServiceName: "integration-test",
					Backends: []Backend{
						{
							Name:   "complete-1",
							Host:   "google.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
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
						{
							Name:   "complete-2",
							Host:   "bing.com",
							Weight: "1",
							Probe: Probe{
								Handler: Handler{
									HTTPGet: &HTTPGetAction{
										Scheme: "http",
										Path:   "/",
										Port:   80,
										Host:   "bing.com",
									},
								},
								TimeoutSeconds: 3,
								PeriodSeconds:  3,
							},
						},
					},
				},
			}
			err = k8sClient.Create(ctx, gslb)
			Expect(err).To(BeNil())

			By("renaming a Gslb backend to a duplicate backend name")

			gslbLookupKey := types.NamespacedName{Name: fooGslbName, Namespace: fooGslbNameSpace}
			err = k8sClient.Get(ctx, gslbLookupKey, gslb)
			Expect(err).To(BeNil())

			gslb.Spec.Backends[1].Name = gslb.Spec.Backends[0].Name
			err = k8sClient.Update(ctx, gslb)
			Expect(string(errors.ReasonForError(err))).Should(Equal("duplicate backend name found: " + gslb.Spec.GetBackends()[0].Name + ". All backend names must be unique"))
		})

	})

})
