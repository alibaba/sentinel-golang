package stat

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseStatNode(t *testing.T) {
	n := NewBaseStatNode(2, 10000)

	assert.Equal(t, uint32(2), n.sampleCount)
	assert.Equal(t, uint32(10000), n.intervalMs)
	assert.Equal(t, int32(0), n.goroutineNum)
	assert.NotNil(t, n.arr)
	assert.NotNil(t, n.metric)
}

func TestBaseStatNode_CurrentGoroutineNum(t *testing.T) {
	n := &BaseStatNode{
		goroutineNum: 1,
	}

	assert.Equal(t, int32(1), n.CurrentGoroutineNum())
}

func TestBaseStatNode_IncreaseGoroutineNum(t *testing.T) {
	n := &BaseStatNode{
		goroutineNum: 1,
	}

	n.IncreaseGoroutineNum()
	assert.Equal(t, int32(2), n.goroutineNum)
}

func TestBaseStatNode_DecreaseGoroutineNum(t *testing.T) {
	n := &BaseStatNode{
		goroutineNum: 1,
	}

	n.DecreaseGoroutineNum()
	assert.Equal(t, int32(0), n.goroutineNum)
}
