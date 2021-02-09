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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CircuitBreakerRule struct {
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Optional
	Id string `json:"id,omitempty"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Required
	Resource string `json:"resource"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=SlowRequestRatio;ErrorRatio;ErrorCount
	// +kubebuilder:validation:Required
	Strategy string `json:"strategy"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	RetryTimeoutMs int32 `json:"retryTimeoutMs"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	MinRequestAmount int64 `json:"minRequestAmount"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	StatIntervalMs int32 `json:"statIntervalMs"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	MaxAllowedRtMs int64 `json:"maxAllowedRtMs"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	Threshold int64 `json:"threshold"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CircuitBreakerRulesSpec defines the desired state of CircuitBreakerRules
type CircuitBreakerRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Type=array
	// +kubebuilder:validation:Optional
	Rules []CircuitBreakerRule `json:"rules"`
}

// CircuitBreakerRulesStatus defines the observed state of CircuitBreakerRules
type CircuitBreakerRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// CircuitBreakerRules is the Schema for the circuitbreakerrules API
type CircuitBreakerRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CircuitBreakerRulesSpec   `json:"spec,omitempty"`
	Status CircuitBreakerRulesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CircuitBreakerRulesList contains a list of CircuitBreakerRules
type CircuitBreakerRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CircuitBreakerRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CircuitBreakerRules{}, &CircuitBreakerRulesList{})
}
