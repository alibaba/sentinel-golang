// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package llm_token_ratelimit

import (
	"fmt"
	"strings"
)

type Rule struct {
	// ID represents the unique ID of the rule (optional).
	ID string `json:"id,omitempty" yaml:"id,omitempty"`
	// Rule resource name, supporting regular expressions, Must be global match.
	Resource string `json:"resource" yaml:"resource"`
	// Rate limiting strategy.
	Strategy Strategy `json:"strategy" yaml:"strategy"`
	// Token encoding method, exclusively for peta rate limiting strategy
	Encoding TokenEncoding `json:"encoding" yaml:"encoding"`
	// Specific rule items
	SpecificItems []*SpecificItem `json:"specificItems" yaml:"specificItems"`
}

func (r *Rule) ResourceName() string {
	if r == nil {
		return "Rule{nil}"
	}
	return r.Resource
}

func (r *Rule) String() string {
	if r == nil {
		return "Rule{nil}"
	}

	var sb strings.Builder
	sb.WriteString("Rule{")

	if r.ID != "" {
		sb.WriteString(fmt.Sprintf("ID:%s, ", r.ID))
	}

	sb.WriteString(fmt.Sprintf("Resource:%s, ", r.Resource))
	sb.WriteString(fmt.Sprintf("Strategy:%s, ", r.Strategy.String()))
	sb.WriteString(fmt.Sprintf("Encoding:%s", r.Encoding.String()))

	if len(r.SpecificItems) > 0 {
		sb.WriteString(", SpecificItems:[")
		for i, item := range r.SpecificItems {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(item.String())
		}
		sb.WriteString("]")
	} else {
		sb.WriteString(", SpecificItems:[]")
	}

	sb.WriteString("}")

	return sb.String()
}

func (r *Rule) setDefaultRuleOption() {
	if r == nil {
		return
	}

	if len(r.Resource) == 0 {
		r.Resource = DefaultResourcePattern
	}

	if len(r.Encoding.Model) == 0 {
		r.Encoding.Model = DefaultTokenEncodingModel[r.Encoding.Provider]
	}

	for idx1, specificItem := range r.SpecificItems {
		if specificItem == nil {
			continue
		}
		if len(specificItem.Identifier.Value) == 0 {
			r.SpecificItems[idx1].Identifier.Value = DefaultIdentifierValuePattern
		}
		for idx2, keyItem := range specificItem.KeyItems {
			if keyItem == nil {
				continue
			}
			if len(keyItem.Key) == 0 {
				r.SpecificItems[idx1].KeyItems[idx2].Key = DefaultKeyPattern
			}
		}
	}
}

func (r *Rule) filterDuplicatedItem() {
	if r == nil {
		return
	}

	occuredKeyItem := make(map[string]struct{})
	var specificItems []*SpecificItem
	for idx1 := len(r.SpecificItems) - 1; idx1 >= 0; idx1-- {
		if r.SpecificItems[idx1] == nil {
			continue
		}
		var keyItems []*KeyItem
		for idx2 := len(r.SpecificItems[idx1].KeyItems) - 1; idx2 >= 0; idx2-- {
			if r.SpecificItems[idx1].KeyItems[idx2] == nil {
				continue
			}
			hash := generateHash(
				r.SpecificItems[idx1].Identifier.String(),
				r.SpecificItems[idx1].KeyItems[idx2].Key,
				r.SpecificItems[idx1].KeyItems[idx2].Token.CountStrategy.String(),
				r.SpecificItems[idx1].KeyItems[idx2].Time.String(),
			)
			if _, exists := occuredKeyItem[hash]; exists {
				continue
			}
			occuredKeyItem[hash] = struct{}{}
			keyItems = append(keyItems, r.SpecificItems[idx1].KeyItems[idx2])
		}
		if len(keyItems) == 0 {
			continue
		}
		specificItems = append(specificItems, &SpecificItem{
			Identifier: r.SpecificItems[idx1].Identifier,
			KeyItems:   keyItems,
		})
	}
	r.SpecificItems = specificItems
}
