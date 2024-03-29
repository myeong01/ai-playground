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

// ContainerSnapshotSpec defines the desired state of ContainerSnapshot
type ContainerSnapshotSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ContainerName  string   `json:"containerName"`
	VersionedNames []string `json:"versionedNames,omitempty"`
}

// ContainerSnapshotStatus defines the observed state of ContainerSnapshot
type ContainerSnapshotStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Snapshots []Snapshot `json:"snapshots,omitempty"`
}

type Snapshot struct {
	Name       string      `json:"name,omitempty"`
	SnapshotAt metav1.Time `json:"snapshotAt,omitempty"`
	CommitId   string      `json:"commitId,omitempty"`
	Status     string      `json:"status,omitempty"`
	Failed     bool        `json:"failed,omitempty"`
	Reason     string      `json:"reason,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// ContainerSnapshot is the Schema for the containersnapshots API
// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ContainerSnapshot struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ContainerSnapshotSpec   `json:"spec,omitempty"`
	Status ContainerSnapshotStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ContainerSnapshotList contains a list of ContainerSnapshot
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ContainerSnapshotList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerSnapshot `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ContainerSnapshot{}, &ContainerSnapshotList{})
}
