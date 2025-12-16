/*
Copyright 2025.

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

// AgenticWorkflowSpec defines the desired state of AgenticWorkflow
type AgenticWorkflowSpec struct {
	// Replicas is the number of workflow engine instances to run
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=10
	// +kubebuilder:default=1
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	// Image is the container image for the workflow engine
	// +kubebuilder:default="merrygoround/agentic-workflow-engine:latest"
	// +optional
	Image string `json:"image,omitempty"`

	// Port is the HTTP server port
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=65535
	// +kubebuilder:default=8080
	// +optional
	Port int32 `json:"port,omitempty"`

	// APIKeys defines the API keys for LLM providers
	// +required
	APIKeys APIKeysSpec `json:"apiKeys"`

	// Resources defines resource requirements for the workflow engine
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// APIKeysSpec defines API keys for LLM providers
type APIKeysSpec struct {
	// OpenAI API key secret reference
	// +required
	OpenAI SecretKeyRef `json:"openai"`

	// Anthropic API key secret reference
	// +required
	Anthropic SecretKeyRef `json:"anthropic"`

	// Google API key secret reference
	// +required
	Google SecretKeyRef `json:"google"`
}

// SecretKeyRef references a key in a Secret
type SecretKeyRef struct {
	// Name is the name of the secret
	// +required
	Name string `json:"name"`

	// Key is the key in the secret
	// +kubebuilder:default="apiKey"
	// +optional
	Key string `json:"key,omitempty"`
}

// ResourceRequirements defines resource limits and requests
type ResourceRequirements struct {
	// CPU resource requirement
	// +kubebuilder:default="100m"
	// +optional
	CPU string `json:"cpu,omitempty"`

	// Memory resource requirement
	// +kubebuilder:default="256Mi"
	// +optional
	Memory string `json:"memory,omitempty"`
}

// AgenticWorkflowStatus defines the observed state of AgenticWorkflow.
type AgenticWorkflowStatus struct {
	// Phase represents the current phase of the workflow
	// Possible values: Pending, Running, Failed, Unknown
	// +optional
	Phase string `json:"phase,omitempty"`

	// AvailableReplicas is the number of available replicas
	// +optional
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ReadyReplicas is the number of ready replicas
	// +optional
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// Endpoint is the service endpoint URL
	// +optional
	Endpoint string `json:"endpoint,omitempty"`

	// conditions represent the current state of the AgenticWorkflow resource.
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
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// AgenticWorkflow is the Schema for the agenticworkflows API
type AgenticWorkflow struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of AgenticWorkflow
	// +required
	Spec AgenticWorkflowSpec `json:"spec"`

	// status defines the observed state of AgenticWorkflow
	// +optional
	Status AgenticWorkflowStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// AgenticWorkflowList contains a list of AgenticWorkflow
type AgenticWorkflowList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []AgenticWorkflow `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AgenticWorkflow{}, &AgenticWorkflowList{})
}
