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

type IsolationRule struct {
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=32
	// +kubebuilder:validation:Optional
	ID string `json:"id,omitempty"`

	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:validation:Required
	Resource string `json:"resource"`

	// +kubebuilder:validation:Type=integer
	// +kubebuilder:validation:Format=int32
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Required
	Threshold int32 `json:"threshold"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// IsolationRulesSpec defines the desired state of IsolationRules
type IsolationRulesSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Type=array
	// +kubebuilder:validation:Optional
	Rules []IsolationRule `json:"rules"`
}

// IsolationRulesStatus defines the observed state of IsolationRules
type IsolationRulesStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// IsolationRules is the Schema for the isolationrules API
type IsolationRules struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   IsolationRulesSpec   `json:"spec,omitempty"`
	Status IsolationRulesStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IsolationRulesList contains a list of IsolationRules
type IsolationRulesList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IsolationRules `json:"items"`
}

func init() {
	SchemeBuilder.Register(&IsolationRules{}, &IsolationRulesList{})
}
