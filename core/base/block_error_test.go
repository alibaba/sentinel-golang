package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlockErrorFromDeepCopy(t *testing.T) {
	bloErr := BlockError{
		blockType:     BlockTypeUnknown,
		blockMsg:      "test",
		rule:          nil,
		snapshotValue: nil,
	}

	newBloErr := NewBlockErrorFromDeepCopy(bloErr)
	assert.True(t, newBloErr != &bloErr)
}
