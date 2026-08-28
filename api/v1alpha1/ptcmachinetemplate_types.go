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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ResourceMap represents a map of resource names to quantities.
type ResourceMap map[corev1.ResourceName]resource.Quantity

// PTCMachineTemplateSpec defines the desired state of PTCMachineTemplate
type PTCMachineTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// The following markers will use OpenAPI v3 schema to validate the value
	// More info: https://book.kubebuilder.io/reference/markers/crd-validation.html

	// foo is an example field of PTCMachineTemplate. Edit ptcmachinetemplate_types.go to remove/update
	// +optional
	//Foo *string `json:"foo,omitempty"`

	// Template defines the PTCMachine spec used to create new machines.
	Template PTCMachineTemplateResource `json:"template"`
}

// PTCMachineTemplateResource describes the data needed to create a PTCMachine from a template.
type PTCMachineTemplateResource struct {
	// ObjectMeta contains metadata for the PTCMachines created from this template (labels, annotations).
	// +optional
	ObjectMeta PTCMachineTemplateResourceObjectMeta `json:"metadata,omitempty"`

	// Spec is the specification of the desired behavior of the PTCMachine.
	Spec PTCMachineSpec `json:"spec"`
}

// PTCMachineTemplateResourceObjectMeta defines metadata allowed on PTCMachines created from templates.
type PTCMachineTemplateResourceObjectMeta struct {
	// Map of string keys and values that can be used to organize and category objects.
	// +optional
	Labels map[string]string `json:"labels,omitempty"`

	// Annotations is an unstructured key value map stored with a resource that may be set by external tools to store and retrieve arbitrary metadata.
	// +optional
	Annotations map[string]string `json:"annotations,omitempty"`
}

// PTCMachineTemplateStatus defines the observed state of PTCMachineTemplate.
type PTCMachineTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// For Kubernetes API conventions, see:
	// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#typical-status-properties

	// conditions represent the current state of the PTCMachineTemplate resource.
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

	// Capacity defines the resource capacity of machines created from this template.
	// +optional
	Capacity ResourceMap `json:"capacity,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:metadata:labels="cluster.x-k8s.io/v1beta1=v1alpha1"
// +kubebuilder:resource:path=ptcmachinetemplates,scope=Namespaced,categories=cluster-api

// PTCMachineTemplate is the Schema for the ptcmachinetemplates API
type PTCMachineTemplate struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PTCMachineTemplate
	// +required
	Spec PTCMachineTemplateSpec `json:"spec"`

	// status defines the observed state of PTCMachineTemplate
	// +optional
	Status PTCMachineTemplateStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PTCMachineTemplateList contains a list of PTCMachineTemplate
type PTCMachineTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PTCMachineTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(func(s *runtime.Scheme) error {
		s.AddKnownTypes(SchemeGroupVersion, &PTCMachineTemplate{}, &PTCMachineTemplateList{})
		return nil
	})
}
