/*

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

package v1alpha

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServiceDefinition describes the service definition
type ServiceDefinition string

const (
	// Internal means that the service is created by the operator
	Internal ServiceDefinition = "Internal"
	// External means that the operator does not create the service, it usues an external one
	External ServiceDefinition = "External"
)

// ContainerDetailsSpec maps the container spec
type ContainerDetailsSpec struct {
	Name            string        `json:"name"`
	Image           string        `json:"image"`
	ImagePullPolicy v1.PullPolicy `json:"imagePullPolicy"`
	ReadinessProbe  CheckProbe    `json:"readinessProbe"`
	LivenessProbe   CheckProbe    `json:"livenessProbe"`
}

// ContainerSpec maps the container spec
type ContainerSpec struct {
	Contaniers ContainerDetailsSpec `json:"contaniers,omitempty"`
}

// TemplateSpec maps the template spec
type TemplateSpec struct {
	Spec ContainerSpec `json:"spec,omitempty"`
}

// CheckProbe contains the "probes" configurations
// as livenessProbe and readinessProbe
type CheckProbe struct {
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	PeriodSeconds       int32 `json:"periodSeconds"`
	TimeoutSeconds      int32 `json:"timeoutSeconds"`
}

// RabbitMQSpec defines the desired state of RabbitMQ
type RabbitMQSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Template v1.StatefulSet `json:"template"`
	Replicas          int32             `json:"replicas"`
	Template          TemplateSpec      `json:"template,omitempty"`
	ServiceDefinition ServiceDefinition `json:"serviceDefinition"`
	ConfigMap         string            `json:"configMap"`
}

// RabbitMQStatus defines the observed state of RabbitMQ
type RabbitMQStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// RabbitMQ is the Schema for the rabbitmqs API
type RabbitMQ struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitMQSpec   `json:"spec,omitempty"`
	Status RabbitMQStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RabbitMQList contains a list of RabbitMQ
type RabbitMQList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RabbitMQ `json:"items"`
}

// NewRabbitMQStruct Init a RabbitMQ struct with default values
func NewRabbitMQStruct() *RabbitMQ {
	return &RabbitMQ{
		Spec:   RabbitMQSpec{},
		Status: RabbitMQStatus{},
	}
}

func init() {
	SchemeBuilder.Register(&RabbitMQ{}, &RabbitMQList{})
}
