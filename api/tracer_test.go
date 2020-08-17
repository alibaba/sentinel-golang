package api

import (
	"errors"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
)

var (
	testRes = base.NewResourceWrapper("a", base.ResTypeCommon, base.Inbound)
)

func TestTraceErrorToEntry(t *testing.T) {
	te := errors.New("biz error")
	ctx := &base.EntryContext{
		Resource: testRes,
		Input:    nil,
	}
	entry := base.NewSentinelEntry(ctx, testRes, nil)
	TraceError(entry, te)
	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, entry.Err(), te)
}
