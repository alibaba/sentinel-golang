package misc

import (
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
)

type RuleCheckSlotMock1 struct {
	name string
}

func (rcs *RuleCheckSlotMock1) Name() string {
	return rcs.name
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
		assert.True(t, slot.Name() == "rcs1")
	})
	assert.True(t, cnt == 1)

	RegisterRuleCheckSlotForResource("abc", rcs2)
	RegisterRuleCheckSlotForResource("abc", rcs3)
	cnt2 := 0
	GetResourceSlotChain("abc").RangeRuleCheckSlot(func(slot base.RuleCheckSlot) {
		cnt2++
	})
	assert.True(t, cnt2 == 3)

	RegisterRuleCheckSlotForResource("abc", rcs4)
	cnt3 := 0
	names := make([]string, 0, 3)
	GetResourceSlotChain("abc").RangeRuleCheckSlot(func(slot base.RuleCheckSlot) {
		cnt3++
		names = append(names, slot.Name())
	})
	assert.True(t, reflect.DeepEqual(names, []string{"rcs1", "rcs2", "rcs3"}))
	assert.True(t, cnt3 == 3)
}
