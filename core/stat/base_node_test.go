package stat

import (
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"

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
	bsn.IncreaseConcurrency()
	assert.Equal(t, int64(1), bsn.MaxConcurrency())
	assert.Equal(t, int64(1), bsn.SecondMaxConcurrency())
	bsn.DecreaseConcurrency()
	assert.Equal(t, int64(1), bsn.MaxConcurrency())
	assert.Equal(t, int64(1), bsn.SecondMaxConcurrency())

	bsn.IncreaseConcurrency()
	bsn.IncreaseConcurrency()
	assert.Equal(t, int64(2), bsn.MaxConcurrency())
	assert.Equal(t, int64(2), bsn.SecondMaxConcurrency())
	bsn.DecreaseConcurrency()
	bsn.DecreaseConcurrency()

	time.Sleep(time.Second * 1)
	bsn.IncreaseConcurrency()
	assert.Equal(t, int64(2), bsn.MaxConcurrency())
	assert.Equal(t, int64(1), bsn.SecondMaxConcurrency())
}

func TestNewCustomizedBaseStatNode(t *testing.T) {
	bsn, err := NewCustomizedBaseStatNode(8, 2000)
	assert.Nil(t, err)
	assert.NotNil(t, bsn)
	bsn.AddCount(base.MetricEventPass, 1)
	assert.Equal(t, int64(1), bsn.GetSum(base.MetricEventPass))
	time.Sleep(time.Second * 2)
	assert.Equal(t, int64(0), bsn.GetSum(base.MetricEventPass))
}
