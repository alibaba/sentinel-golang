package isolation

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/stretchr/testify/assert"
)

func Test_Isolation_Pass(t *testing.T) {
	slot := &Slot{}
	statSlot := stat.Slot{}
	res := base.NewResourceWrapper("abc", base.ResTypeCommon, base.Inbound)
	resNode := stat.GetOrCreateResourceNode("abc", base.ResTypeCommon)
	ctx := &base.EntryContext{
		Resource: res,
		StatNode: resNode,
		Input: &base.SentinelInput{
			BatchCount: 1,
		},
		RuleCheckResult: nil,
		Data:            nil,
	}

	ret := slot.Check(ctx)
	if ret != nil {
		t.Fail()
		return
	}

	r := &Rule{
		Resource:   "abc",
		MetricType: Concurrency,
		Threshold:  10,
	}
	_, e := LoadRules([]*Rule{r})
	if e != nil {
		t.Fail()
		return
	}
	for i := 0; i < 10; i++ {
		ret = slot.Check(ctx)
		if ret != nil {
			t.Fail()
			return
		}
		statSlot.OnEntryPassed(ctx)
	}

	ret = slot.Check(ctx)
	assert.NotNil(t, ret)
}

func Test_RegexIsolation_Pass(t *testing.T) {
	slot := &Slot{}
	statSlot1 := stat.Slot{}
	statSlot2 := stat.Slot{}
	res1 := base.NewResourceWrapper("abc/123", base.ResTypeCommon, base.Inbound)
	res2 := base.NewResourceWrapper("abc/456", base.ResTypeCommon, base.Inbound)
	resNode1 := stat.GetOrCreateResourceNode("abc/123", base.ResTypeCommon)
	resNode2 := stat.GetOrCreateResourceNode("abc/456", base.ResTypeCommon)
	ctx1 := &base.EntryContext{
		Resource: res1,
		StatNode: resNode1,
		Input: &base.SentinelInput{
			BatchCount: 1,
		},
		RuleCheckResult: nil,
		Data:            nil,
	}
	ctx2 := &base.EntryContext{
		Resource: res2,
		StatNode: resNode2,
		Input: &base.SentinelInput{
			BatchCount: 1,
		},
		RuleCheckResult: nil,
		Data:            nil,
	}

	ret1 := slot.Check(ctx1)
	if ret1 != nil {
		t.Fail()
		return
	}
	ret2 := slot.Check(ctx2)
	if ret2 != nil {
		t.Fail()
		return
	}

	r := &Rule{
		Resource:   "abc/\\d+",
		MetricType: Concurrency,
		Threshold:  10,
		Regex:      true,
	}
	_, e := LoadRules([]*Rule{r})
	if e != nil {
		t.Fail()
		return
	}

	n := 10
	for i := 0; i < n; i++ {
		ret := slot.Check(ctx1)
		if ret != nil {
			t.Fail()
			return
		}

		statSlot1.OnEntryPassed(ctx1)
	}

	ret1 = slot.Check(ctx1)
	assert.NotNil(t, ret1)

	for i := 0; i < n; i++ {
		statSlot1.OnCompleted(ctx1)
	}

	for i := 0; i < n; i++ {
		ret := slot.Check(ctx1)
		if ret != nil {
			t.Fail()
			return
		}
		statSlot1.OnEntryPassed(ctx1)
	}

	for i := 0; i < n; i++ {
		ret := slot.Check(ctx2)
		if ret != nil {
			t.Fail()
			return
		}
		statSlot2.OnEntryPassed(ctx2)
	}
	ret2 = slot.Check(ctx2)
	assert.NotNil(t, ret2)

	for i := 0; i < n; i++ {
		statSlot2.OnCompleted(ctx2)
	}

	for i := 0; i < n; i++ {
		ret := slot.Check(ctx2)
		if ret != nil {
			t.Fail()
			return
		}
		statSlot2.OnEntryPassed(ctx2)
	}
}
