package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var flag = 0

func exitHandlerMock(entry *SentinelEntry, ctx *EntryContext) error {
	flag += 1
	return nil
}

func TestSentinelEntry_WhenExit(t *testing.T) {
	flag = 0
	sc := NewSlotChain()
	ctx := sc.GetPooledContext()
	entry := NewSentinelEntry(ctx, nil, sc)
	entry.WhenExit(exitHandlerMock)
	entry.Exit()
	assert.True(t, flag == 1)

	entry.Exit()
	assert.True(t, flag == 1)
}
