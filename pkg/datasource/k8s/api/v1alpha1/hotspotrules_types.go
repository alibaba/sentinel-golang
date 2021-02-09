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

type SpecificValue struct {
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=KindInt;KindString;KindBool;KindFloat64
	// +kubebuilder:validation:Required
	ValKind string `json:"valKind"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Required
	ValStr string `json:"valStr"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Required
	Threshold int64 `json:"threshold"`
}

type HotspotRule struct {
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
	// +kubebuilder:validation:Enum=Concurrency;QPS
	// +kubebuilder:validation:Required
	MetricType string `json:"metricType"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Enum=Reject;Throttling
	// +kubebuilder:validation:Optional
	ControlBehavior string `json:"controlBehavior"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	ParamIndex int32 `json:"paramIndex"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	Threshold int64 `json:"threshold"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	MaxQueueingTimeMs int64 `json:"maxQueueingTimeMs"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	BurstCount int64 `json:"burstCount"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	DurationInSec int64 `json:"durationInSec"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int64
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Optional
	ParamsMaxCapacity int64 `json:"paramsMaxCapacity"`

	// +kubebuilder:validation:Optional
	SpecificItems []SpecificValue `json:"specificItems"`
}

// HotspotRulesSpec defines the desired state of HotspotRules
type HotspotRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Type=array
	// +kubebuilder:validation:Optional
	Rules []HotspotRule `json:"rules"`
}

// HotspotRulesStatus defines the observed state of HotspotRules
type HotspotRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// HotspotRules is the Schema for the hotspotrules API
type HotspotRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   HotspotRulesSpec   `json:"spec,omitempty"`
	Status HotspotRulesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// HotspotRulesList contains a list of HotspotRules
type HotspotRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []HotspotRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&HotspotRules{}, &HotspotRulesList{})
}
