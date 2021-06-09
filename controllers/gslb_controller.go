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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	gslbv1alpha1 "github.com/snapp-cab/gslb-controller/api/v1alpha1"
)

// GslbReconciler reconciles a Gslb object
type GslbReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const gslbFinalizer = "gslb.snappcloud.io/finalizer"

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

	// Lookup the route instance for this reconcile request
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
		log.Error(err, "Failed to get Routes")
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
	}

	// Add finalizer for this CR
	if !controllerutil.ContainsFinalizer(gslb, gslbFinalizer) {
		controllerutil.AddFinalizer(gslb, gslbFinalizer)
		err = r.Update(ctx, gslb)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	// Check if the service already exists, if not create a new one
	currentGslb, err := GetGslb(ctx, gslb)
	if err != nil && err.Error() == "NotFound" {
		log.Info("Creating a new gslb service", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
		err = CreateGslb(ctx, gslb)
		if err != nil {
			log.Error(err, "Failed to create new gslb service", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
			return ctrl.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get gslb service")
		return ctrl.Result{}, err
	}

	// Ensure the gslb service status with the desired state
	if !reflect.DeepEqual(currentGslb, gslb) {
		log.Info("Updating the gslb service", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
		err := UpdateGslb(ctx, gslb)
		if err != nil {
			log.Error(err, "Failed to update the gslb service", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GslbReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&gslbv1alpha1.Gslb{}).
		Complete(r)
}

func (r *GslbReconciler) finalizeGslb(ctx context.Context, reqLogger logr.Logger, gslb *gslbv1alpha1.Gslb) error {
	reqLogger.Info("Deleting the gslb service", "Gslb.Namespace", gslb.Namespace, "Gslb.Name", gslb.Name)
	err := DeleteGslb(ctx, gslb)
	if err != nil {
		return err
	}
	reqLogger.Info("Successfully deleted and finalized gslb")
	return nil
}
