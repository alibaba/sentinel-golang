package core

type Context struct {
	ResWrapper  *ResourceWrapper
	Entry       *CtxEntry
	Node        node
	Count       uint64
	Input       *SentinelInput
	Output      *SentinelOutput
	FeatureData map[interface{}]interface{}
}

func NewContext() *Context {
	ctx := &Context{
		Input:  newInput(),
		Output: newOutput(),
	}
	ctx.Input.Context = ctx
	ctx.Output.Context = ctx
	return ctx
}

type SentinelInput struct {
	Context *Context
	// store some values in this context when calling context in slot.
	data map[interface{}]interface{}
}

func newInput() *SentinelInput {
	return &SentinelInput{}
}

type SentinelOutput struct {
	Context     *Context
	CheckResult *RuleCheckResult
	msg         string
	// store output data.
	data map[interface{}]interface{}
}

func newOutput() *SentinelOutput {
	return &SentinelOutput{}
}

// Reset init Context,
func (ctx *Context) Reset() {
	// reset all fields of ctx
	ctx.ResWrapper = nil
	ctx.Entry = nil
	ctx.Node = nil
	ctx.Count = 0
	ctx.Input = nil
	ctx.Output = nil
	ctx.FeatureData = nil
}
