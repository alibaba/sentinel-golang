package stat

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"

	"github.com/stretchr/testify/assert"
)

func TestGetAllNodes(t *testing.T) {
	ctx := base.NewEmptyEntryContext()
	ctx.Input = &base.SentinelInput{}
	ctx.Resource = &base.ResourceWrapper{}
	data := GetAllNodes(ctx)
	assert.Nil(t, data)
}

func TestIsMonitorBlocked(t *testing.T) {
	assert.False(t, IsMonitorBlocked(nil))

	ctx := base.NewEmptyEntryContext()
	ctx.Input = &base.SentinelInput{}
	ctx.Data = make(map[interface{}]interface{})
	assert.False(t, IsMonitorBlocked(ctx))

	PutOutputAttachment(ctx, KeyIsMonitorBlocked, true)
	assert.True(t, IsMonitorBlocked(ctx))

	PutOutputAttachment(ctx, KeyIsMonitorBlocked, false)
	assert.False(t, IsMonitorBlocked(ctx))

	PutOutputAttachment(ctx, KeyIsMonitorBlocked, 1)
	assert.False(t, IsMonitorBlocked(ctx))
}

func mockCtx() *base.EntryContext {
	ctx := base.NewEmptyEntryContext()
	ctx.Input = &base.SentinelInput{}
	ctx.Data = make(map[interface{}]interface{})
	return ctx
}

func TestSetFirstMonitorNode(t *testing.T) {
	//nil
	SetFirstMonitorNodeAndRule(nil, nil, nil)
	_, _, isBlocked := GetMonitorBlockInfo(nil)
	assert.False(t, isBlocked)

	// first set
	ctx := mockCtx()
	resourceNode := NewResourceNode("test", base.ResTypeRPC)
	rule := &base.MockRule{Id: "1"}
	SetFirstMonitorNodeAndRule(ctx, resourceNode, rule)
	gotNode, gotRule, isBlocked2 := GetMonitorBlockInfo(ctx)

	assert.True(t, isBlocked2)
	assert.Equal(t, gotNode, resourceNode)
	assert.Equal(t, gotRule, rule)

	// second set

	resourceNode2 := NewResourceNode("test2", base.ResTypeRPC)
	rule2 := &base.MockRule{Id: "2"}
	SetFirstMonitorNodeAndRule(ctx, resourceNode2, rule2)
	gotNode2, gotRule2, isBlock2 := GetMonitorBlockInfo(ctx)

	assert.True(t, isBlock2)
	assert.Equal(t, gotNode2, resourceNode)
	assert.Equal(t, gotRule2, rule)
}
