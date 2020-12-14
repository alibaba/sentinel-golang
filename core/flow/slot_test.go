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
