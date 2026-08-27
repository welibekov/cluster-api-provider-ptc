/*
Copyright 2026.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// NetworkSpec defines shared network configuration across cluster and machine specs.
type NetworkSpec struct {
	// Name specifies the network/VLAN name on PTC (e.g., vlan-2415).
	Name string `json:"name"`

	// Subnet specifies the expected (e.g.,255.255.255.0).
	// +optional
	Subnet string `json:"subnet,omitempty"`

	// IPFromPoolRef references an InClusterIPPool for dynamic static IP allocation.
	// +optional
	IPFromPoolRef *corev1.TypedLocalObjectReference `json:"ipFromPoolRef,omitempty"`
}

// PTCClusterSpec defines the desired state of PTCCluster
type PTCClusterSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of PTCCluster. Edit ptccluster_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`

	// The PTC Region the cluster lives in.
	Region string `json:"region"`

	// NetworkSpec encapsulates all things related to PTC network.
	// +optional
	Network NetworkSpec `json:"network"`

	// IdentityRef points to the Secret containing PTC credentials.
	// +optional
	IdentityRef *corev1.SecretReference `json:"identityRef,omitempty"`

	// ControlPlaneEndpoint represents the endpoint for the cluster API server.
	// +optional
	ControlPlaneEndpoint clusterv1.APIEndpoint `json:"controlPlaneEndpoint,omitempty"`
}

// PTCClusterStatus defines the observed state of PTCCluster.
type PTCClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PTCCluster resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	Ready               bool `json:"ready"`
	InfrastructureReady bool `json:"infrastructureReady"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// PTCCluster is the Schema for the ptcclusters API
type PTCCluster struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PTCCluster
	// +required
	Spec PTCClusterSpec `json:"spec"`

	// status defines the observed state of PTCCluster
	// +optional
	Status PTCClusterStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PTCClusterList contains a list of PTCCluster
type PTCClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PTCCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PTCCluster{}, &PTCClusterList{})
		return nil
	})
}
