package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEntryContext_IsBlocked(t *testing.T) {
	ctx := NewEmptyEntryContext()
	assert.False(t, ctx.IsBlocked(), "empty context with no result should indicate pass")
	ctx.RuleCheckResult = NewTokenResultBlocked(BlockTypeUnknown)
	assert.True(t, ctx.IsBlocked(), "context with blocked request should indicate blocked")
}
