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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
)

const (
	gslbContentFinalizer = "gslb.snappcloud.io/gslbcontent-finalizer"
)

// GslbContentReconciler reconciles a GslbContent object
type GslbContentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=gslb.snappcloud.io,resources=gslbcontents,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=gslb.snappcloud.io,resources=gslbcontents/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=gslb.snappcloud.io,resources=gslbcontents/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GslbContent object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile
func (r *GslbContentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Lookup the gslbcontent instance for this reconcile request
	gslbcon := &gslbv1alpha1.GslbContent{}
	err := r.Get(ctx, req.NamespacedName, gslbcon)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get GslbContent")
		return ctrl.Result{}, err
	}
	// Check if the gslb instance is marked to be deleted, which is
	// indicated by the deletion timestamp being set.
	if gslbcon.GetDeletionTimestamp() != nil {
		// Run finalization logic for gslbContentFinalizer. If the
		// finalization logic fails, don't remove the finalizer so
		// that we can retry during the next reconciliation.
		if err := r.finalizeGslbcon(ctx, log, gslbcon); err != nil {
			return ctrl.Result{}, err
		}

		// Remove gslbContentFinalizer. Once all finalizers have been
		// removed, the object will be deleted.
		controllerutil.RemoveFinalizer(gslbcon, gslbContentFinalizer)
		err := r.Update(ctx, gslbcon) // TODO: this again triggers a reconcile, duplicate: causes: Resource not found. Ignoring since object must be delete
		if err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(gslbcon, gslbContentFinalizer) {
		controllerutil.AddFinalizer(gslbcon, gslbContentFinalizer)
		log.Info("Add finalizer to gslbcontent", "GslbContent.Name", gslbcon.Name)
		err = r.Update(ctx, gslbcon) // TODO shouldn't we reconcile it? the update itself will trigger a reconcile
		if err != nil {
			return ctrl.Result{}, err
		}

	}

	// Sync the gslbcontent with driver
	log.Info("Sync gslbcontent with driver", "GslbContent.Name", gslbcon.Name)
	err = CreateGslbcon(ctx, gslbcon)
	if err != nil {
		log.Error(err, "Failed to sync gslbcontent with driver", "GslbContent.Name", gslbcon.Name)
		return ctrl.Result{}, err
	}
	log.Info("Succesfully synced gslbcontent with driver")
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GslbContentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gslbv1alpha1.GslbContent{}).
		Complete(r)
}

func (r *GslbContentReconciler) finalizeGslbcon(ctx context.Context, reqLogger logr.Logger, gslbcon *gslbv1alpha1.GslbContent) error {
	reqLogger.Info("Deleting the gslbcontent with driver", "GslbContent.Name", gslbcon.Name)
	err := DeleteGslbcon(ctx, gslbcon)
	if err != nil {
		reqLogger.Error(err, "Failed to delete the gslbcontent", "GslbContent.Name", gslbcon.Name)
		return err
	}
	reqLogger.Info("Successfully deleted and finalized gslbcontent")
	return nil
}
