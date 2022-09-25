/*
Copyright 2022.

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

// NodeRequestSpec defines the desired state of NodeRequest
type NodeRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of NodeRequest. Edit noderequest_types.go to remove/update
	Foo string `json:"foo,omitempty"`
}

// NodeRequestStatus defines the observed state of NodeRequest
type NodeRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	NodeName    string `json:"NodeName"`
	NodeCpu     string `json:"NodeCpu"`
	NodeCpuRate string `json:"NodeCpuRate"`
	NodeMem     string `json:"NodeMem"`
	NodeMemRate string `json:"NodeMemRate"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// NodeRequest is the Schema for the noderequests API
type NodeRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NodeRequestSpec   `json:"spec,omitempty"`
	Status NodeRequestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NodeRequestList contains a list of NodeRequest
type NodeRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NodeRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NodeRequest{}, &NodeRequestList{})
}
