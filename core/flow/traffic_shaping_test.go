package flow

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
)

func TestTrafficShapingController_Rule(t *testing.T) {
	type fields struct {
		flowCalculator TrafficShapingCalculator
		flowChecker    TrafficShapingChecker
		rule           *FlowRule
	}

	rule := []byte(`{"resource":"Sentinel","limitApp":"","grade":1,"count":0,"strategy":0,"controlBehavior":0,"warmUpPeriodSec":10,"maxQueueingTimeMs":1000,"clusterMode":true,"clusterConfig":{"thresholdType":1}}`)
	flowRule := &FlowRule{}
	_ = json.Unmarshal(rule, &flowRule)

	tests := []struct {
		name   string
		fields fields
		want   *FlowRule
	}{
		{"test1", fields{nil, nil, flowRule}, flowRule},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traffic := &TrafficShapingController{
				flowCalculator: tt.fields.flowCalculator,
				flowChecker:    tt.fields.flowChecker,
				rule:           tt.fields.rule,
			}
			if got := traffic.Rule(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrafficShapingController.Rule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrafficShapingController_FlowChecker(t *testing.T) {
	type fields struct {
		flowCalculator TrafficShapingCalculator
		flowChecker    TrafficShapingChecker
		rule           *FlowRule
	}

	flowChecker := NewThrottlingChecker(1000)

	tests := []struct {
		name   string
		fields fields
		want   TrafficShapingChecker
	}{
		{"test1", fields{nil, flowChecker, nil}, flowChecker},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traffic := &TrafficShapingController{
				flowCalculator: tt.fields.flowCalculator,
				flowChecker:    tt.fields.flowChecker,
				rule:           tt.fields.rule,
			}
			if got := traffic.FlowChecker(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrafficShapingController.FlowChecker() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrafficShapingController_PerformChecking(t *testing.T) {
	type fields struct {
		flowCalculator TrafficShapingCalculator
		flowChecker    TrafficShapingChecker
		rule           *FlowRule
	}
	type args struct {
		node         base.StatNode
		acquireCount uint32
		flag         int32
	}

	flowCalculator := NewDefaultTrafficShapingCalculator(99)
	flowChecker := NewThrottlingChecker(1000)
	tokenResult := base.NewTokenResultPass()
	node := stat.NewBaseStatNode(1, 1000)

	tests := []struct {
		name   string
		fields fields
		args   args
		want   *base.TokenResult
	}{
		{"test1", fields{flowCalculator, flowChecker, nil}, args{}, tokenResult},
		{"test2", fields{flowCalculator, flowChecker, nil}, args{node, 1, 1000}, tokenResult},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traffic := &TrafficShapingController{
				flowCalculator: tt.fields.flowCalculator,
				flowChecker:    tt.fields.flowChecker,
				rule:           tt.fields.rule,
			}
			if got := traffic.PerformChecking(tt.args.node, tt.args.acquireCount, tt.args.flag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrafficShapingController.PerformChecking() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTrafficShapingController_FlowCalculator(t *testing.T) {
	type fields struct {
		flowCalculator TrafficShapingCalculator
		flowChecker    TrafficShapingChecker
		rule           *FlowRule
	}

	flowCalculator := NewDefaultTrafficShapingCalculator(99)

	tests := []struct {
		name   string
		fields fields
		want   TrafficShapingCalculator
	}{
		{"test1", fields{flowCalculator, nil, nil}, flowCalculator},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			traffic := &TrafficShapingController{
				flowCalculator: tt.fields.flowCalculator,
				flowChecker:    tt.fields.flowChecker,
				rule:           tt.fields.rule,
			}
			if got := traffic.FlowCalculator(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("TrafficShapingController.FlowCalculator() = %v, want %v", got, tt.want)
			}
		})
	}
}
