package core

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEntryContext_IsBlocked(t *testing.T) {
	ctx := NewEmptyEntryContext()
	assert.False(t, ctx.IsBlocked(), "empty context with no result should indicate pass")
	ctx.Output = &SentinelOutput{LastResult: NewTokenResultBlocked(BlockTypeUnknown, "")}
	assert.True(t, ctx.IsBlocked(), "context with blocked request should indicate blocked")
}
