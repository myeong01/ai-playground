/*
Copyright 2023.

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

// DynamicMountSpec defines the desired state of DynamicMount
type DynamicMountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	DatasetName    string `json:"datasetName"`
	ContainerName  string `json:"containerName"`
	MountPath      string `json:"mountPath,omitempty"`
	DatasetSubPath string `json:"datasetSubPath,omitempty"`
	ReadOnly       bool   `json:"readOnly,omitempty"`
}

// DynamicMountStatus defines the observed state of DynamicMount
type DynamicMountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Mounted bool   `json:"mounted,omitempty"`
	Failed  bool   `json:"failed,omitempty"`
	Reason  string `json:"reason,omitempty"`
	Status  string `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DynamicMount is the Schema for the dynamicmounts API
type DynamicMount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DynamicMountSpec   `json:"spec,omitempty"`
	Status DynamicMountStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DynamicMountList contains a list of DynamicMount
type DynamicMountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DynamicMount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DynamicMount{}, &DynamicMountList{})
}
