package flow

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/stretchr/testify/assert"
)

func Test_FlowSlot_StandaloneStat(t *testing.T) {
	slot := &Slot{}
	statSLot := &StandaloneStatSlot{}
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

	slot.Check(ctx)

	r1 := &Rule{
		Resource:               "abc",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		// Use standalone statistic, using single-bucket-sliding-windows
		StatIntervalInMs: 20000,
		Threshold:        100,
		RelationStrategy: CurrentResource,
	}
	_, e := LoadRules([]*Rule{r1})
	if e != nil {
		logging.Error(e, "")
		t.Fail()
		return
	}

	for i := 0; i < 50; i++ {
		ret := slot.Check(ctx)
		if ret != nil {
			t.Fail()
			return
		}
		statSLot.OnEntryPassed(ctx)
	}
	assert.True(t, getTrafficControllerListFor("abc")[0].boundStat.readOnlyMetric.GetSum(base.MetricEventPass) == 50)
}
