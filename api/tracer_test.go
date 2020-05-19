package api

import (
	"errors"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/stretchr/testify/assert"
)

var (
	testRes = base.NewResourceWrapper("a", base.ResTypeCommon, base.Inbound)
)

func TestTraceErrorToCtx(t *testing.T) {
	type args struct {
		ctx   *base.EntryContext
		err   error
		count TraceErrorOption
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "TestTraceErrorToCtx",
			args: args{
				ctx:   nil,
				err:   nil,
				count: WithCount(10),
			},
			want: 10,
		},
	}
	testStatNode := stat.NewResourceNode("a", base.ResTypeCommon)
	tests[0].args.ctx = &base.EntryContext{
		Resource: testRes,
		StatNode: testStatNode,
		Input:    nil,
	}
	tests[0].args.err = errors.New("biz error")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TraceErrorToCtx(tt.args.ctx, tt.args.err, tt.args.count)
			assert.True(t, tt.args.ctx.StatNode.GetSum(base.MetricEventError) == tt.want)
		})
	}
}

func TestTraceErrorToEntry(t *testing.T) {
	type args struct {
		entry *base.SentinelEntry
		err   error
		count TraceErrorOption
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			name: "TestTraceErrorToEntry",
			args: args{
				entry: nil,
				err:   nil,
				count: WithCount(10),
			},
			want: 10,
		},
	}

	testStatNode := stat.NewResourceNode("a", base.ResTypeCommon)
	ctx := &base.EntryContext{
		Resource: testRes,
		StatNode: testStatNode,
		Input:    nil,
	}
	tests[0].args.entry = base.NewSentinelEntry(ctx, testRes, nil)
	tests[0].args.err = errors.New("biz error")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TraceErrorToEntry(tt.args.entry, tt.args.err, tt.args.count)
			time.Sleep(time.Millisecond * 10)
			assert.True(t, ctx.StatNode.GetSum(base.MetricEventError) == tt.want)
		})
	}
}
