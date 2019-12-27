package log

import (
	"github.com/sentinel-group/sentinel-golang/core/base"
)

type LogSlot struct {
}

func (s *LogSlot) OnEntryPassed(_ *base.EntryContext) {

}

func (s *LogSlot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	// TODO: write sentinel-block.log here
	panic("implement me")
}

func (s *LogSlot) OnCompleted(_ *base.EntryContext) {

}
