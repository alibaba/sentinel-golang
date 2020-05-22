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
	type args struct {
		entry *base.SentinelEntry
		err   error
	}
	te := errors.New("biz error")
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "TestTraceErrorToEntry",
			args: args{
				entry: nil,
				err:   nil,
			},
			want: te,
		},
	}

	ctx := &base.EntryContext{
		Resource: testRes,
		Input:    nil,
	}
	tests[0].args.entry = base.NewSentinelEntry(ctx, testRes, nil)
	tests[0].args.err = te

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TraceError(tt.args.entry, tt.args.err)
			time.Sleep(time.Millisecond * 10)
			assert.Equal(t, tests[0].args.entry.Context().Err(), tt.want)
		})
	}
}
