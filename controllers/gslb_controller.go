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
	"reflect"

	"github.com/go-logr/logr"
	"github.com/m-yosefpor/utils"
	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// GslbReconciler reconciles a Gslb object
type GslbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const (
	gslbFinalizer = "gslb.snappcloud.io/gslb-finalizer"
	prefix        = "gslb"
)

//+kubebuilder:rbac:groups=gslb.snappcloud.io,resources=gslbs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gslb.snappcloud.io,resources=gslbs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gslb.snappcloud.io,resources=gslbs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Gslb object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *GslbReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Lookup the gslb instance for this reconcile request
	gslb := &gslbv1alpha1.Gslb{}
	err := r.Get(ctx, req.NamespacedName, gslb)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Gslb")
		return ctrl.Result{}, err
	}
	// Check if the gslb instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	if gslb.GetDeletionTimestamp() != nil {
		// Run finalization logic for gslbFinalizer. If the
		// finalization logic fails, don't remove the finalizer so
		// that we can retry during the next reconciliation.
		if err := r.finalizeGslb(ctx, log, gslb); err != nil {
			return ctrl.Result{}, err
		}

		// Remove gslbFinalizer. Once all finalizers have been
		// removed, the object will be deleted.
		controllerutil.RemoveFinalizer(gslb, gslbFinalizer)
		err := r.Update(ctx, gslb)
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(gslb, gslbFinalizer) {
		log.Info("Add finalizer to gslb", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
		controllerutil.AddFinalizer(gslb, gslbFinalizer)
		err = r.Update(ctx, gslb)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	var found *gslbv1alpha1.GslbContent
	var desiredGslbconNames []string
	for _, b := range gslb.Spec.Backends {
		// TODO: we can continue to add other backends and only requeue after the for loop
		// so at least we will add good backends and only backends with issue remains

		// Define a new gslbcontent
		gslbCon, err := r.createGslbconFromGsblbackend(gslb, b)
		if err != nil {
			log.Error(err, "Error creating GslbContent from Gslb", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "Gslb.Backend", b.Name)
			return ctrl.Result{}, err
		}
		desiredGslbconNames = append(desiredGslbconNames, gslbCon.Name)
		// Check if Gslbcontents already exist for this Gslb, if not create missing ones
		found = &gslbv1alpha1.GslbContent{}
		// TODO: get by predictive name can be enhanced to get the list of gslbcon who has label/ownerRef of gslbUID
		// and then loop over them to match a field such as backendName on them
		// e.g.
		// gslbContentList := &gslbv1alpha1.GslbContentList{}
		// listOpts := []client.ListOption{
		// 	client.MatchingLabels(labelsForGslbcon(string(gslb.GetUID()))),
		// }

		// if err = r.List(ctx, gslbContentList, listOpts...); err != nil {
		// 	log.Error(err, "Failed to list gslbContents", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
		// 	return ctrl.Result{}, err
		// }
		err = r.Get(ctx, types.NamespacedName{Name: gslbCon.Name, Namespace: ""}, found)
		if err != nil && errors.IsNotFound(err) {

			log.Info("Creating a new GslbContent", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", gslbCon.Name)
			err = r.Create(ctx, gslbCon)
			if err != nil {
				log.Error(err, "Failed to create new GslbContent", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", gslbCon.Name)
				return ctrl.Result{}, err
			}
			// GslbContent created successfully - return and requeue
			continue
		} else if err != nil {
			log.Error(err, "Failed to get GslbContent")
			return ctrl.Result{}, err
		}

		// If Gslbcontents already exist, check if it is deeply equal with desrired state
		if !reflect.DeepEqual(gslbCon.Spec, found.Spec) {
			log.Info("Updating GslbContent", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", gslbCon.Name)
			found.Spec = gslbCon.Spec
			err := r.Update(ctx, found)
			if err != nil {
				log.Error(err, "Failed to update GslbContent", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", gslbCon.Name)
				return ctrl.Result{}, err
			}
		}
	}

	// Check gslbContents which refrence this Gslb, and delete extra GslbContents
	gslbContentList := &gslbv1alpha1.GslbContentList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(labelsForGslbcon(gslb.Name, gslb.Namespace)),
	}

	if err = r.List(ctx, gslbContentList, listOpts...); err != nil {
		log.Error(err, "Failed to list gslbContents", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
		return ctrl.Result{}, err
	}

	// Delete gslbcon if is not in desired list
	for _, g := range gslbContentList.Items {
		if !utils.IsIn(g.Name, desiredGslbconNames) {
			log.Info("Deleting GslbContent", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", g.Name)
			err = r.Delete(ctx, &g)
			if err != nil {
				log.Error(err, "Failed to delete extra gslbContents", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", g.Name)
				return ctrl.Result{}, err
			}
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GslbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gslbv1alpha1.Gslb{}).
		// Watches all Gslbcontents and trigger reconcile on the corresponding gslb
		// since cluster-scoped gslbcontent can not be owned by namespaced gslb resource
		Watches(
			&source.Kind{Type: &gslbv1alpha1.GslbContent{}}, handler.EnqueueRequestsFromMapFunc(func(obj client.Object) []reconcile.Request {
				reconcileRequests := []reconcile.Request{{
					NamespacedName: types.NamespacedName{
						Name:      obj.GetLabels()["gslbName"],
						Namespace: obj.GetLabels()["gslbNamespace"],
					},
				}}
				return reconcileRequests
			}),
		).
		// Owns(&gslbv1alpha1.GslbContent{}). // Cause automatic GC: if gslb object deleted, all gslbcons will be deleted too
		Complete(r)
}

func (r *GslbReconciler) finalizeGslb(ctx context.Context, log logr.Logger, gslb *gslbv1alpha1.Gslb) error {
	log.Info("Deleting the gslb", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
	// Check gslbContents which refrence this Gslb, and delete them
	gslbContentList := &gslbv1alpha1.GslbContentList{}
	listOpts := []client.ListOption{
		client.MatchingLabels(labelsForGslbcon(gslb.Name, gslb.Namespace)),
	}

	if err := r.List(ctx, gslbContentList, listOpts...); err != nil {
		log.Error(err, "Failed to list gslbContents", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
		return err
	}

	// Delete gslbcon if is not in desired list
	for _, g := range gslbContentList.Items {
		log.Info("Deleting GslbContent", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", g.Name)
		err := r.Delete(ctx, &g)
		if err != nil {
			log.Error(err, "Failed to delete gslbContents", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name, "GslbContent.Name", g.Name)
			return err
		}
	}

	log.Info("Successfully deleted and finalized gslb")
	return nil
}

// labelsForGslbcon returns the labels for selecting the resources
// belonging to the given gslb CR name.
func labelsForGslbcon(gslbName, gslbNamespace string) map[string]string {
	return map[string]string{"gslbName": gslbName, "gslbNamespace": gslbNamespace}
}

func (r *GslbReconciler) createGslbconFromGsblbackend(gslb *gslbv1alpha1.Gslb, b gslbv1alpha1.Backend) (*gslbv1alpha1.GslbContent, error) {
	gslbUID := string(gslb.GetUID())
	ls := labelsForGslbcon(gslb.Name, gslb.Namespace)

	gslbcon := &gslbv1alpha1.GslbContent{
		ObjectMeta: metav1.ObjectMeta{
			Name:   prefix + "-" + gslbUID + "-" + b.Name,
			Labels: ls,
		},
		Spec: gslbv1alpha1.GslbContentSpec{
			ServiceName: gslb.Spec.ServiceName,
			Backend:     b,
		},
	}
	// Set Gslb instance as the owner and controller of GslbContent
	// cluster-scoped resource must not have a namespace-scoped owner
	// err := ctrl.SetControllerReference(gslb, gslbcon, r.Scheme)
	return gslbcon, nil
}
