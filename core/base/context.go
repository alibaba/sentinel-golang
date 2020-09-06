package base

import "github.com/alibaba/sentinel-golang/util"

type EntryContext struct {
	entry *SentinelEntry
	// internal error when sentinel Entry or
	// biz error of downstream
	err error
	// Use to calculate RT
	startTime uint64
	// the rt of this transaction
	rt uint64

	Resource *ResourceWrapper
	StatNode StatNode

	Input *SentinelInput
	// the result of rule slots check
	RuleCheckResult *TokenResult
	// reserve for storing some intermediate data from the Entry execution process
	Data map[interface{}]interface{}
}

func (ctx *EntryContext) SetEntry(entry *SentinelEntry) {
	ctx.entry = entry
}

func (ctx *EntryContext) Entry() *SentinelEntry {
	return ctx.entry
}

func (ctx *EntryContext) Err() error {
	return ctx.err
}

func (ctx *EntryContext) SetError(err error) {
	ctx.err = err
}

func (ctx *EntryContext) StartTime() uint64 {
	return ctx.startTime
}

func (ctx *EntryContext) IsBlocked() bool {
	if ctx.RuleCheckResult == nil {
		return false
	}
	return ctx.RuleCheckResult.IsBlocked()
}

func (ctx *EntryContext) PutRt(rt uint64) {
	ctx.rt = rt
}

func (ctx *EntryContext) Rt() uint64 {
	if ctx.rt == 0 {
		rt := util.CurrentTimeMillis() - ctx.StartTime()
		return rt
	}
	return ctx.rt
}

func NewEmptyEntryContext() *EntryContext {
	return &EntryContext{}
}

// The input data of sentinel
type SentinelInput struct {
	AcquireCount uint32
	Flag         int32
	Args         []interface{}
	// store some values in this context when calling context in slot.
	Attachments map[interface{}]interface{}
}

func (i *SentinelInput) reset() {
	i.AcquireCount = 1
	i.Flag = 0
	if len(i.Args) != 0 {
		i.Args = make([]interface{}, 0)
	}
	if len(i.Attachments) != 0 {
		i.Attachments = make(map[interface{}]interface{})
	}
}

// Reset init EntryContext,
func (ctx *EntryContext) Reset() {
	// reset all fields of ctx
	ctx.entry = nil
	ctx.err = nil
	ctx.startTime = 0
	ctx.rt = 0
	ctx.Resource = nil
	ctx.StatNode = nil
	ctx.Input.reset()
	if ctx.RuleCheckResult == nil {
		ctx.RuleCheckResult = NewTokenResultPass()
	} else {
		ctx.RuleCheckResult.ResetToPass()
	}
	if len(ctx.Data) != 0 {
		ctx.Data = make(map[interface{}]interface{})
	}
}
