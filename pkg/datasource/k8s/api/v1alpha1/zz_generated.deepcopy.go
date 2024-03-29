//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CircuitBreakerRule) DeepCopyInto(out *CircuitBreakerRule) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CircuitBreakerRule.
func (in *CircuitBreakerRule) DeepCopy() *CircuitBreakerRule {
	if in == nil {
		return nil
	}
	out := new(CircuitBreakerRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CircuitBreakerRules) DeepCopyInto(out *CircuitBreakerRules) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CircuitBreakerRules.
func (in *CircuitBreakerRules) DeepCopy() *CircuitBreakerRules {
	if in == nil {
		return nil
	}
	out := new(CircuitBreakerRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CircuitBreakerRules) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CircuitBreakerRulesList) DeepCopyInto(out *CircuitBreakerRulesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CircuitBreakerRules, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CircuitBreakerRulesList.
func (in *CircuitBreakerRulesList) DeepCopy() *CircuitBreakerRulesList {
	if in == nil {
		return nil
	}
	out := new(CircuitBreakerRulesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CircuitBreakerRulesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CircuitBreakerRulesSpec) DeepCopyInto(out *CircuitBreakerRulesSpec) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]CircuitBreakerRule, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CircuitBreakerRulesSpec.
func (in *CircuitBreakerRulesSpec) DeepCopy() *CircuitBreakerRulesSpec {
	if in == nil {
		return nil
	}
	out := new(CircuitBreakerRulesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CircuitBreakerRulesStatus) DeepCopyInto(out *CircuitBreakerRulesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CircuitBreakerRulesStatus.
func (in *CircuitBreakerRulesStatus) DeepCopy() *CircuitBreakerRulesStatus {
	if in == nil {
		return nil
	}
	out := new(CircuitBreakerRulesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowRule) DeepCopyInto(out *FlowRule) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowRule.
func (in *FlowRule) DeepCopy() *FlowRule {
	if in == nil {
		return nil
	}
	out := new(FlowRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowRules) DeepCopyInto(out *FlowRules) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowRules.
func (in *FlowRules) DeepCopy() *FlowRules {
	if in == nil {
		return nil
	}
	out := new(FlowRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlowRules) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowRulesList) DeepCopyInto(out *FlowRulesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]FlowRules, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowRulesList.
func (in *FlowRulesList) DeepCopy() *FlowRulesList {
	if in == nil {
		return nil
	}
	out := new(FlowRulesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *FlowRulesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowRulesSpec) DeepCopyInto(out *FlowRulesSpec) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]FlowRule, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowRulesSpec.
func (in *FlowRulesSpec) DeepCopy() *FlowRulesSpec {
	if in == nil {
		return nil
	}
	out := new(FlowRulesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *FlowRulesStatus) DeepCopyInto(out *FlowRulesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new FlowRulesStatus.
func (in *FlowRulesStatus) DeepCopy() *FlowRulesStatus {
	if in == nil {
		return nil
	}
	out := new(FlowRulesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HotspotRule) DeepCopyInto(out *HotspotRule) {
	*out = *in
	if in.SpecificItems != nil {
		in, out := &in.SpecificItems, &out.SpecificItems
		*out = make([]SpecificValue, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HotspotRule.
func (in *HotspotRule) DeepCopy() *HotspotRule {
	if in == nil {
		return nil
	}
	out := new(HotspotRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HotspotRules) DeepCopyInto(out *HotspotRules) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HotspotRules.
func (in *HotspotRules) DeepCopy() *HotspotRules {
	if in == nil {
		return nil
	}
	out := new(HotspotRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HotspotRules) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HotspotRulesList) DeepCopyInto(out *HotspotRulesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]HotspotRules, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HotspotRulesList.
func (in *HotspotRulesList) DeepCopy() *HotspotRulesList {
	if in == nil {
		return nil
	}
	out := new(HotspotRulesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *HotspotRulesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HotspotRulesSpec) DeepCopyInto(out *HotspotRulesSpec) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]HotspotRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HotspotRulesSpec.
func (in *HotspotRulesSpec) DeepCopy() *HotspotRulesSpec {
	if in == nil {
		return nil
	}
	out := new(HotspotRulesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *HotspotRulesStatus) DeepCopyInto(out *HotspotRulesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new HotspotRulesStatus.
func (in *HotspotRulesStatus) DeepCopy() *HotspotRulesStatus {
	if in == nil {
		return nil
	}
	out := new(HotspotRulesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IsolationRule) DeepCopyInto(out *IsolationRule) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IsolationRule.
func (in *IsolationRule) DeepCopy() *IsolationRule {
	if in == nil {
		return nil
	}
	out := new(IsolationRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IsolationRules) DeepCopyInto(out *IsolationRules) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IsolationRules.
func (in *IsolationRules) DeepCopy() *IsolationRules {
	if in == nil {
		return nil
	}
	out := new(IsolationRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IsolationRules) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IsolationRulesList) DeepCopyInto(out *IsolationRulesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]IsolationRules, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IsolationRulesList.
func (in *IsolationRulesList) DeepCopy() *IsolationRulesList {
	if in == nil {
		return nil
	}
	out := new(IsolationRulesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *IsolationRulesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IsolationRulesSpec) DeepCopyInto(out *IsolationRulesSpec) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]IsolationRule, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IsolationRulesSpec.
func (in *IsolationRulesSpec) DeepCopy() *IsolationRulesSpec {
	if in == nil {
		return nil
	}
	out := new(IsolationRulesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *IsolationRulesStatus) DeepCopyInto(out *IsolationRulesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new IsolationRulesStatus.
func (in *IsolationRulesStatus) DeepCopy() *IsolationRulesStatus {
	if in == nil {
		return nil
	}
	out := new(IsolationRulesStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SpecificValue) DeepCopyInto(out *SpecificValue) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SpecificValue.
func (in *SpecificValue) DeepCopy() *SpecificValue {
	if in == nil {
		return nil
	}
	out := new(SpecificValue)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemRule) DeepCopyInto(out *SystemRule) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemRule.
func (in *SystemRule) DeepCopy() *SystemRule {
	if in == nil {
		return nil
	}
	out := new(SystemRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemRules) DeepCopyInto(out *SystemRules) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemRules.
func (in *SystemRules) DeepCopy() *SystemRules {
	if in == nil {
		return nil
	}
	out := new(SystemRules)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemRules) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemRulesList) DeepCopyInto(out *SystemRulesList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SystemRules, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemRulesList.
func (in *SystemRulesList) DeepCopy() *SystemRulesList {
	if in == nil {
		return nil
	}
	out := new(SystemRulesList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SystemRulesList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemRulesSpec) DeepCopyInto(out *SystemRulesSpec) {
	*out = *in
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]SystemRule, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemRulesSpec.
func (in *SystemRulesSpec) DeepCopy() *SystemRulesSpec {
	if in == nil {
		return nil
	}
	out := new(SystemRulesSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SystemRulesStatus) DeepCopyInto(out *SystemRulesStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SystemRulesStatus.
func (in *SystemRulesStatus) DeepCopy() *SystemRulesStatus {
	if in == nil {
		return nil
	}
	out := new(SystemRulesStatus)
	in.DeepCopyInto(out)
	return out
}
