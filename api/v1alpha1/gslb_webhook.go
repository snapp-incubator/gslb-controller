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
	"fmt"
	"sync"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var (
	gslblog             = logf.Log.WithName("gslb-resource")
	claimedServiceNames sync.Map
)

func (r *Gslb) SetupWebhookWithManager(mgr ctrl.Manager) error {

	err := ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
	if err != nil {
		return err
	}
	clientOptions := client.Options{
		Scheme: mgr.GetScheme(),
		Mapper: mgr.GetRESTMapper(),
	}
	c, err := client.New(mgr.GetConfig(), clientOptions)
	if err != nil {
		return fmt.Errorf("failed to create new client: %v", err)
	}

	// Filling the ClaimedServiceNames on webhook initialization.
	// After initialization, the map update will happen on object creation.
	gslbList := GslbList{}
	err = c.List(context.TODO(), &gslbList, &client.ListOptions{})
	if err != nil {
		return fmt.Errorf("unable to list gslbs: %v", err)
	}
	for _, gslb := range gslbList.Items {
		claimedServiceNames.Store(gslb.Spec.ServiceName, struct{}{})
	}
	return err
}

//+kubebuilder:webhook:path=/mutate-gslb-snappcloud-io-v1alpha1-gslb,mutating=true,failurePolicy=fail,sideEffects=None,groups=gslb.snappcloud.io,resources=gslbs,verbs=create;update,versions=v1alpha1,name=mgslb.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Defaulter = &Gslb{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *Gslb) Default() {
	gslblog.Info("default", "name", r.Name)
}

//+kubebuilder:webhook:path=/validate-gslb-snappcloud-io-v1alpha1-gslb,mutating=false,failurePolicy=fail,sideEffects=None,groups=gslb.snappcloud.io,resources=gslbs,verbs=create;update;delete,versions=v1alpha1,name=vgslb.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &Gslb{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *Gslb) ValidateCreate() error {
	gslblog.Info("validate create", "name", r.Name)
	err := r.validateGslb()
	if err != nil {
		return err
	}

	// validate the serviceName is unique
	if _, loaded := claimedServiceNames.LoadOrStore(r.Spec.ServiceName, struct{}{}); loaded {
		return fmt.Errorf("'%v' serviceName is already claimed. please try another serviceName", r.Spec.ServiceName)
	}
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *Gslb) ValidateUpdate(old runtime.Object) error {
	gslblog.Info("validate update", "name", r.Name)
	err := r.validateGslb()
	if err != nil {
		return err
	}
	oldGslb, ok := old.(*Gslb)
	if !ok {
		return fmt.Errorf("failed to convert object to gslb type")
	}
	if r.Spec.ServiceName != oldGslb.Spec.ServiceName {
		if _, ok := claimedServiceNames.Load(r.Spec.ServiceName); ok {
			return fmt.Errorf("'%v' serviceName is already claimed. please try another serviceName", r.Spec.ServiceName)
		}
		claimedServiceNames.Store(r.Spec.ServiceName, struct{}{})
		claimedServiceNames.Delete(oldGslb.Spec.ServiceName)
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *Gslb) ValidateDelete() error {
	gslblog.Info("validate delete", "name", r.Name)
	claimedServiceNames.Delete(r.Spec.ServiceName)
	return nil
}

func (r *Gslb) validateGslb() error {
	// validate no repetitive backend name
	visited := make(map[string]struct{}, len(r.Spec.Backends))
	for _, b := range r.Spec.Backends {
		if _, exists := visited[b.Name]; exists {
			return fmt.Errorf("duplicate backend name found: %v. All backend names must be unique", b.Name)
		}
		visited[b.Name] = struct{}{}
	}
	return nil
}
