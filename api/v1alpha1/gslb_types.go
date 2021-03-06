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
	ServiceName ServiceName `json:"serviceName"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinItems:=1
	// +kubebuilder:validation:MaxItems:=10
	Backends []Backend `json:"backends"`
}

func (g *GslbSpec) GetBackends() []Backend {
	if g.Backends != nil {
		return g.Backends
	}
	return []Backend{}
}

// ServiceName for Gslb. The fullname will be ServiceName.service.ha
// +kubebuilder:validation:Required
// +kubebuilder:validation:MinLength:=1
// +kubebuilder:validation:MaxLength:=50
type ServiceName string

type Backend struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	// +kubebuilder:validation:MaxLength:=50
	Name string `json:"name"`
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength:=1
	// DNS1123SubdomainMaxLength = 253
	// +kubebuilder:validation:MaxLength:=253
	/// kubebuilder:validation:Format:="hostname"
	/// kubebuilder:validation:Format:="ipv4"
	Host string `json:"host"`
	// +kubebuilder:validation:Optional
	// +kubebuilder:default:="1"
	Weight string `json:"weight,omitempty"`
	// +kubebuilder:validation:Optional
	Probe Probe `json:"probe,omitempty"`
}

// GslbStatus defines the observed state of Gslb
type GslbStatus struct {
	// +kubebuilder:default:="Pending"
	// +kubebuilder:validation:Enum:=Pending;Configured;Degraded;serviceNameAlreadyClaimed
	State  string `json:"state,omitempty"`
	Reason string `json:"reason,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// +kubebuilder:printcolumn:name="SERVICENAME",type=string,JSONPath=`.spec.serviceName`,description="Service Name of the GSLB"
// +kubebuilder:printcolumn:name="STATE",type=string,JSONPath=`.status.state`,description="State of GSLB"
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
