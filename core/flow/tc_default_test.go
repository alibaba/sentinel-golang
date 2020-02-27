package flow

import (
	"reflect"
	"testing"

	"github.com/sentinel-group/sentinel-golang/core/base"
	"github.com/sentinel-group/sentinel-golang/core/stat"
)

func TestDefaultTrafficShapingCalculator_CalculateAllowedTokens(t *testing.T) {
	type fields struct {
		threshold float64
	}
	type args struct {
		in0 base.StatNode
		in1 uint32
		in2 int32
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   float64
	}{
		{"test1", fields{0x88888877373}, args{}, 0x88888877373},
		{"test2", fields{88888877373}, args{}, 88888877373},
		{"test3", fields{0}, args{}, 0},
		{"test4", fields{-1.9999}, args{}, -1.9999},
		{"test5", fields{-88888877373}, args{}, -88888877373},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultTrafficShapingCalculator{
				threshold: tt.fields.threshold,
			}
			if got := d.CalculateAllowedTokens(tt.args.in0, tt.args.in1, tt.args.in2); got != tt.want {
				t.Errorf("DefaultTrafficShapingCalculator.CalculateAllowedTokens() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTrafficShapingChecker_DoCheck(t *testing.T) {
	type fields struct {
		metricType MetricType
	}
	type args struct {
		node         base.StatNode
		acquireCount uint32
		threshold    float64
	}

	tokenResult := base.NewTokenResultPass()
	tokenResult0 := base.NewTokenResultBlocked(base.BlockTypeFlow, "Flow")

	node2_0 := stat.NewBaseStatNode(1, 1000)
	node2_1 := stat.NewBaseStatNode(2, 1000)
	node2_1.IncreaseGoroutineNum() // set goroutineNum add 1

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *base.TokenResult
	}{
		// node = nil test
		{"test1_0", fields{Concurrency}, args{nil, 10000, 10000}, tokenResult},
		{"test1_1", fields{QPS}, args{nil, 10000, 10000}, tokenResult},

		// node !=nil && MetricType == Concurrency && CurCount == 0
		{"test2_0", fields{Concurrency}, args{node2_0, 10000, 10000}, tokenResult},
		// node !=nil && MetricType == QPS && CurCount == 0
		{"test2_1", fields{QPS}, args{node2_0, 10000, 10000}, tokenResult},
		// node !=nil && MetricType == Concurrency && CurCount == 1
		{"test2_2", fields{Concurrency}, args{node2_1, 10000, 10000}, tokenResult0},
		// other condition node !=nil && MetricType == QPS && GetQPS(?) > 0
		// also return tokenResult0
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &DefaultTrafficShapingChecker{
				metricType: tt.fields.metricType,
			}
			if got := d.DoCheck(tt.args.node, tt.args.acquireCount, tt.args.threshold); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DefaultTrafficShapingChecker.DoCheck() = %v, want %v", got, tt.want)
			}
		})
	}
}
