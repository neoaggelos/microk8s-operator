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

type AddonRepositorySpec struct {
	// Name is the name used to refer to the addon repository.
	Name string `json:"name"`
	// Repository is the source to use for the addon repository.
	Repository string `json:"repository"`
	// Reference is the a git tag to checkout (leave empty to fetch the default branch).
	Reference string `json:"reference,omitempty"`
}

// ConfigurationSpec defines the desired state of Configuration
type ConfigurationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// AddonRepositories is the list of addon repositories to configure.
	AddonRepositories []AddonRepositorySpec `json:"addonRepositories,omitempty"`

	// ContainerdRegistryConfigs is configuration for the image registries.
	// The key name is the name of the registry, and the value is the contents
	// of the hosts.toml file.
	ContainerdRegistryConfigs map[string]string `json:"containerdRegistryConfigs,omitempty"`

	// PodCIDR is the CIDR to use for pods. This should match any CNI configuration.
	PodCIDR string `json:"podCIDR,omitempty"`

	// ExtraSANs is a list of extra subject alternative names to add to the server certificates.
	ExtraSANs []string `json:"extraSANs,omitempty"`

	// ExtraSANIPs is a list of extra IP addresses to include as SANs to the server certificates.
	ExtraSANIPs []string `json:"extraSANIPs,omitempty"`

	// ExtraKubeletArgs is a list of extra arguments to pass to kubelet.
	ExtraKubeletArgs []string `json:"extraKubeletArgs,omitempty"`
}

type AddonRepositoryStatus struct {
	// Name is the name of the addon repository
	Name string `json:"name"`
	// Status is the status of the addon repository
	Status string `json:"status"`
}

// ConfigurationStatus defines the observed state of Configuration
type ConfigurationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// AddonRepositories is the status of the addon repositories
	AddonRepositories []AddonRepositoryStatus `json:"addonRepositories"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:scope=Cluster

// Configuration is the Schema for the configurations API
type Configuration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigurationSpec   `json:"spec,omitempty"`
	Status ConfigurationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ConfigurationList contains a list of Configuration
type ConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Configuration `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Configuration{}, &ConfigurationList{})
}
