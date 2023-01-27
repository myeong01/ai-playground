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
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ClusterRoleSpec defines the desired state of ClusterRole
type ClusterRoleSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ClusterRole. Edit clusterrole_types.go to remove/update
	IsApproved            bool                `json:"isApproved,omitempty"`
	ParentClusterRoleName string              `json:"parentClusterRoleName,omitempty"`
	Rules                 []rbacv1.PolicyRule `json:"rules"`
}

// ClusterRoleStatus defines the observed state of ClusterRole
type ClusterRoleStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ClusterRoleName string              `json:"clusterRoleName,omitempty"`
	Rules           []rbacv1.PolicyRule `json:"rules,omitempty"`
	IsFailed        bool                `json:"isFailed,omitempty"`
	Reason          string              `json:"reason,omitempty"`
	IsChildChecked  bool                `json:"isChildChecked"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// ClusterRole is the Schema for the clusterroles API
type ClusterRole struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterRoleSpec   `json:"spec,omitempty"`
	Status ClusterRoleStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterRoleList contains a list of ClusterRole
type ClusterRoleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterRole `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterRole{}, &ClusterRoleList{})
}
