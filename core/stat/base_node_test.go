package stat

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	// timespan of per slot
	BucketLengthInMs uint32 = 500
	// the number of slots
	SampleCount uint32 = 20
	// interval(ms) of sliding window, 10s
	IntervalInMs uint32 = 10 * 1000
)

func TestBaseStatNodeGoroutineNum(t *testing.T) {
	bsn := NewBaseStatNode(SampleCount, IntervalInMs)
	bsn.IncreaseGoroutineNum()
	assert.Equal(t, int64(1), bsn.MaxConcurrency())
	assert.Equal(t, int64(1), bsn.SecondMaxConcurrency())
	bsn.DecreaseGoroutineNum()
	assert.Equal(t, int64(1), bsn.MaxConcurrency())
	assert.Equal(t, int64(1), bsn.SecondMaxConcurrency())

	bsn.IncreaseGoroutineNum()
	bsn.IncreaseGoroutineNum()
	assert.Equal(t, int64(2), bsn.MaxConcurrency())
	assert.Equal(t, int64(2), bsn.SecondMaxConcurrency())
	bsn.DecreaseGoroutineNum()
	bsn.DecreaseGoroutineNum()

	time.Sleep(time.Second * 1)
	bsn.IncreaseGoroutineNum()
	assert.Equal(t, int64(2), bsn.MaxConcurrency())
	assert.Equal(t, int64(1), bsn.SecondMaxConcurrency())
}
