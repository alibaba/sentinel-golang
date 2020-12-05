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

package misc

import (
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/stretchr/testify/assert"
)

type RuleCheckSlotMock1 struct {
	name string
}

func (rcs *RuleCheckSlotMock1) Name() string {
	return rcs.name
}

func (rcs *RuleCheckSlotMock1) Order() uint32 {
	return 0
}

func (rcs *RuleCheckSlotMock1) Check(ctx *base.EntryContext) *base.TokenResult {
	return nil
}

func TestRegisterRuleCheckSlotForResource(t *testing.T) {
	rcs1 := &RuleCheckSlotMock1{
		name: "rcs1",
	}
	rcs2 := &RuleCheckSlotMock1{
		name: "rcs2",
	}
	rcs3 := &RuleCheckSlotMock1{
		name: "rcs3",
	}
	rcs4 := &RuleCheckSlotMock1{
		name: "rcs3",
	}

	assert.True(t, GetResourceSlotChain("abc0") == nil)
	RegisterRuleCheckSlotForResource("abc", rcs1)
	assert.True(t, GetResourceSlotChain("abc") != nil)
	cnt := 0
	GetResourceSlotChain("abc").RangeRuleCheckSlot(func(slot base.RuleCheckSlot) {
		cnt++
	})
	assert.True(t, cnt == 2)

	RegisterRuleCheckSlotForResource("abc", rcs2)
	RegisterRuleCheckSlotForResource("abc", rcs3)
	cnt2 := 0
	GetResourceSlotChain("abc").RangeRuleCheckSlot(func(slot base.RuleCheckSlot) {
		cnt2++
	})
	assert.True(t, cnt2 == 4)

	RegisterRuleCheckSlotForResource("abc", rcs4)
	cnt3 := 0
	names := make([]string, 0, 4)
	GetResourceSlotChain("abc").RangeRuleCheckSlot(func(slot base.RuleCheckSlot) {
		cnt3++
		names = append(names, slot.Name())
	})
	assert.True(t, reflect.DeepEqual(names, []string{"rcs1", "rcs2", "rcs3", system.RuleCheckSlotName}))
	assert.True(t, cnt3 == 4)
}
