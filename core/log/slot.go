package log

import (
	"github.com/alibaba/sentinel-golang/core/base"
)

const (
	StatSlotName = "sentinel-core-log-stat-slot"
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Name() string {
	return StatSlotName
}

func (s *Slot) OnEntryPassed(_ *base.EntryContext) {

}

func (s *Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// TODO: write sentinel-block.log here
}

func (s *Slot) OnCompleted(_ *base.EntryContext) {

}
