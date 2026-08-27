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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NfsProvisionerFinalizer is the finalizer applied to NfsProvisioner resources
// by its managing controller.
const NfsProvisionerFinalizer = "nfsprovisioner.gcore-sfs-controller.io"

// DefaultAPITokenSecretKey is the Secret key used when
// spec.apiTokenSecretRef does not specify one.
const DefaultAPITokenSecretKey = "apiToken"

// NfsProvisionerSpec defines the desired state of NfsProvisioner
type NfsProvisionerSpec struct {
	// APIToken is the API token used to authenticate with Gcore Cloud.
	//
	// Deprecated: storing the token in the custom resource exposes it to
	// anyone who can read the resource (kubectl get, etcd backups, audit
	// logs). Use APITokenSecretRef instead.
	// +optional
	APIToken string `json:"apiToken,omitempty"`

	// APITokenSecretRef references a key of a Secret in the same namespace
	// as the NfsProvisioner that holds the Gcore Cloud API token.
	// If the key is not specified, it defaults to "apiToken".
	// Exactly one of APIToken and APITokenSecretRef must be set.
	// +optional
	APITokenSecretRef *corev1.SecretKeySelector `json:"apiTokenSecretRef,omitempty"`
	// APIURL is the URL of the Gcore Cloud API.
	// +optional
	APIURL string `json:"apiURL,omitempty"`

	// File share region ID
	RegionID int `json:"region"`

	// File share project ID
	ProjectID int `json:"project"`

	// Provisioner helm repository
	// +optional
	HelmRepository string `json:"helmRepository,omitempty"`

	// Provisioner Helm chart name
	// +optional
	ChartName string `json:"chartName,omitempty"`

	// Provisioner Helm chart version
	// +optional
	ChartVersion string `json:"chartVersion,omitempty"`

	// Provisioner image version
	// +optional
	ImageVersion string `json:"imageVersion,omitempty"`

	// Paused can be used to prevent controllers from processing the Provisioner and all its associated objects.
	// +optional
	Paused bool `json:"paused"`
}

// NfsProvisionerStatus defines the observed state of NfsProvisioner
type NfsProvisionerStatus struct {
	// Ready denotes that all nfs file share provisioners has been deployed and running
	ProvisionersReady bool `json:"provisionersReady"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// NfsProvisioner is the Schema for the nfsprovisioners API
type NfsProvisioner struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NfsProvisionerSpec   `json:"spec,omitempty"`
	Status NfsProvisionerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// NfsProvisionerList contains a list of NfsProvisioner
type NfsProvisionerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NfsProvisioner `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NfsProvisioner{}, &NfsProvisionerList{})
}
