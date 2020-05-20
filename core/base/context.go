package base

type EntryContext struct {
	// Use to calculate RT
	startTime uint64

	Resource *ResourceWrapper
	StatNode StatNode

	Input *SentinelInput
	// the result of rule slots check
	RuleCheckResult *TokenResult
	// reserve for storing some intermediate data from the Entry execution process
	Data map[interface{}]interface{}
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

func newEmptyInput() *SentinelInput {
	return &SentinelInput{
		AcquireCount: 1,
		Flag:         0,
		Args:         make([]interface{}, 0, 0),
		Attachments:  make(map[interface{}]interface{}),
	}
}

// Reset init EntryContext,
func (ctx *EntryContext) Reset() {
	// reset all fields of ctx
	ctx.startTime = 0
	ctx.Resource = nil
	ctx.StatNode = nil
	ctx.Input = nil
	if ctx.RuleCheckResult == nil {
		ctx.RuleCheckResult = NewTokenResultPass()
	} else {
		ctx.RuleCheckResult.ResetToPass()
	}
	ctx.Data = nil
}
