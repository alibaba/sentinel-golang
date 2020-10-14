package flow

import (
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

const (
	StatSlotName = "sentinel-core-flow-standalone-stat-slot"
)

var (
	DefaultStandaloneStatSlot = &StandaloneStatSlot{}
)

type StandaloneStatSlot struct {
}

func (s *StandaloneStatSlot) Name() string {
	return StatSlotName
}

func (s StandaloneStatSlot) OnEntryPassed(ctx *base.EntryContext) {
	res := ctx.Resource.Name()
	for _, tc := range getTrafficControllerListFor(res) {
		if !tc.boundStat.reuseResourceStat {
			if tc.boundStat.writeOnlyMetric != nil {
				tc.boundStat.writeOnlyMetric.AddCount(base.MetricEventPass, int64(ctx.Input.BatchCount))
			} else {
				logging.Error(errors.New("nil independent write statistic"), "flow module: nil statistic for traffic control", "rule", tc.rule)
			}
		}
	}
}

func (s StandaloneStatSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// Do nothing
}

func (s StandaloneStatSlot) OnCompleted(ctx *base.EntryContext) {
	// Do nothing
}
