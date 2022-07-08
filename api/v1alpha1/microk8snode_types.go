/*
Copyright 2022 Angelos Kolaitis.

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

// MicroK8sNodeStatus defines the observed state of MicroK8sNode
type MicroK8sNodeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// LastUpdate is the timestamp of the last update of this node.
	LastUpdate metav1.Time `json:"lastUpdate"`

	// Revision is the installed MicroK8s snap revision.
	Revision string `json:"revision"`

	// Channel is the channel MicroK8s is tracking.
	Channel string `json:"channel"`

	// Version is the MicroK8s snap version.
	Version string `json:"version"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster
// +kubebuilder:printcolumn:name="LastUpdate",type="date",JSONPath=".status.lastUpdate",description="age"
// +kubebuilder:printcolumn:name="Version",type="string",JSONPath=".status.version",description="Installed version"
// +kubebuilder:printcolumn:name="Revision",type="string",JSONPath=".status.revision",description="Installed revision"
// +kubebuilder:printcolumn:name="Channel",type="string",JSONPath=".status.channel",description="Tracking channel"
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="age"

// MicroK8sNode is the Schema for the microk8snodes API
type MicroK8sNode struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Status MicroK8sNodeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true
//

// MicroK8sNodeList contains a list of MicroK8sNode
type MicroK8sNodeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MicroK8sNode `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MicroK8sNode{}, &MicroK8sNodeList{})
}
