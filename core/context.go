package core

type EntryContext struct {
	ResWrapper  *ResourceWrapper
	Entry       *CtxEntry
	StatNode    Node
	Count       uint64
	Input       *SentinelInput
	Output      *SentinelOutput
	FeatureData map[interface{}]interface{}
}

func NewEntryContext() *EntryContext {
	ctx := &EntryContext{
		Input:  newInput(),
		Output: newOutput(),
	}
	ctx.Input.Context = ctx
	ctx.Output.Context = ctx
	return ctx
}

type SentinelInput struct {
	Context *EntryContext
	// store some values in this context when calling context in slot.
	data map[interface{}]interface{}
}

func newInput() *SentinelInput {
	return &SentinelInput{}
}

type SentinelOutput struct {
	Context     *EntryContext
	CheckResult *RuleCheckResult
	msg         string
	// store output data.
	data map[interface{}]interface{}
}

func newOutput() *SentinelOutput {
	return &SentinelOutput{}
}

// Reset init EntryContext,
func (ctx *EntryContext) Reset() {
	// reset all fields of ctx
	ctx.ResWrapper = nil
	ctx.Entry = nil
	ctx.StatNode = nil
	ctx.Count = 0
	ctx.Input = nil
	ctx.Output = nil
	ctx.FeatureData = nil
}
