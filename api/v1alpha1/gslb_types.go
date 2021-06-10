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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GslbSpec defines the desired state of Gslb
type GslbSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// ServiceName for Gslb. The fullname will be ServiceName.service.ha
	// +kubebuilder:validation:Required
	ServiceName string `json:"serviceName,omitempty"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxItems:=10
	Backends []Backend `json:"backends,omitempty"`
}

type Backend struct {
	// +kubebuilder:validation:Required
	Name string `json:"name,omitempty"`
	// +kubebuilder:validation:Required
	Host string `json:"host,omitempty"`
	// +kubebuilder:validation:Required
	Weight string `json:"weight,omitempty"`
	// +kubebuilder:validation:Optional
	Probe Probe `json:"probe,omitempty"`
}

// GslbStatus defines the observed state of Gslb
type GslbStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Gslb is the Schema for the gslbs API
type Gslb struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GslbSpec   `json:"spec,omitempty"`
	Status            GslbStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GslbList contains a list of Gslb
type GslbList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Gslb `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Gslb{}, &GslbList{})
}
