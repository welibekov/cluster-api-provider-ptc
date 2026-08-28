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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PTCMachineSpec defines the desired state of PTCMachine
type PTCMachineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of PTCMachine. Edit ptcmachine_types.go to remove/update
	// +optional
	Foo *string `json:"foo,omitempty"`

	// +optional
	InstanceType string `json:"instanceType"`

	// +optional
	BootDiskSize int `json:"bootDiskSize"`

	// +optional
	Image string `json:"image"`

	// +optional
	Tags []string `json:"tags"`

	// +optional
	SSHKey string `json:"sshKey"`

	// +optional
	Network NetworkSpec `json:"network"`

	ProviderID *string `json:"providerID,omitempty"`
}

// PTCMachineStatus defines the observed state of PTCMachine.
type PTCMachineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PTCMachine resource.
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

	// InstanceID is the unique identifier assigned by PTC Cloud to the underlying VM.
	// +optional
	InstanceID string `json:"instanceID,omitempty"`

	// Addresses contains the PTC VM's associated network addresses (Internal/External IPs, Hostname).
	// +optional
	Addresses []clusterv1.MachineAddress `json:"addresses,omitempty"`

	// Ready denotes that the PTC VM infrastructure is fully provisioned and ready.
	// +optional
	Ready bool `json:"ready"`

	// FailureMessage provides a descriptive explanation of the terminal error.
	// +optional
	FailureMessage *string `json:"failureMessage,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1alpha1"

// PTCMachine is the Schema for the ptcmachines API
type PTCMachine struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PTCMachine
	// +required
	Spec PTCMachineSpec `json:"spec"`

	// status defines the observed state of PTCMachine
	// +optional
	Status PTCMachineStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PTCMachineList contains a list of PTCMachine
type PTCMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PTCMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PTCMachine{}, &PTCMachineList{})
		return nil
	})
}
