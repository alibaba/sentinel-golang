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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type FlowRule struct {
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
	// +kubebuilder:validation:Enum=Direct;WarmUp
	// +kubebuilder:validation:Optional
	TokenCalculateStrategy string `json:"tokenCalculateStrategy"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Reject;Throttling
	// +kubebuilder:validation:Optional
	ControlBehavior string `json:"controlBehavior"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	Threshold int64 `json:"threshold"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=CurrentResource;AssociatedResource
	// +kubebuilder:validation:Optional
	RelationStrategy string `json:"relationStrategy"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Optional
	RefResource string `json:"refResource"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Optional
	MaxQueueingTimeMs int32 `json:"maxQueueingTimeMs"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	WarmUpPeriodSec int32 `json:"warmUpPeriodSec"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:validation:Optional
	WarmUpColdFactor int32 `json:"warmUpColdFactor"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	StatIntervalInMs int32 `json:"statIntervalInMs"`
}

// FlowRulesSpec defines the desired state of FlowRules
type FlowRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Type=array
	// +kubebuilder:validation:Optional
	Rules []FlowRule `json:"rules"`
}

// FlowRulesStatus defines the observed state of FlowRules
type FlowRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// FlowRules is the Schema for the flowrules API
type FlowRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FlowRulesSpec   `json:"spec,omitempty"`
	Status FlowRulesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// FlowRulesList contains a list of FlowRules
type FlowRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FlowRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FlowRules{}, &FlowRulesList{})
}
