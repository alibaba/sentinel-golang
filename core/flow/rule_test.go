package flow

import (
	"encoding/json"
	"testing"
)

// ResourceName Case sensitive
func TestFlowRule_ResourceName(t *testing.T) {
	type fields struct {
		ID                uint64
		Resource          string
		LimitOrigin       string
		MetricType        MetricType
		Count             float64
		RelationStrategy  RelationStrategy
		ControlBehavior   ControlBehavior
		RefResource       string
		WarmUpPeriodSec   uint32
		MaxQueueingTimeMs uint32
		ClusterMode       bool
		ClusterConfig     ClusterRuleConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"test1", fields{0, "Sentinel", "", QPS, 0, Direct, Reject, "", 10, 1000, true, ClusterRuleConfig{AvgLocalThreshold}}, "Sentinel",
		}, {
			"test2", fields{0, "sentinel", "", QPS, 0, Direct, Reject, "", 10, 1000, true, ClusterRuleConfig{AvgLocalThreshold}}, "sentinel",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FlowRule{
				ID:                tt.fields.ID,
				Resource:          tt.fields.Resource,
				LimitOrigin:       tt.fields.LimitOrigin,
				MetricType:        tt.fields.MetricType,
				Count:             tt.fields.Count,
				RelationStrategy:  tt.fields.RelationStrategy,
				ControlBehavior:   tt.fields.ControlBehavior,
				RefResource:       tt.fields.RefResource,
				WarmUpPeriodSec:   tt.fields.WarmUpPeriodSec,
				MaxQueueingTimeMs: tt.fields.MaxQueueingTimeMs,
				ClusterMode:       tt.fields.ClusterMode,
				ClusterConfig:     tt.fields.ClusterConfig,
			}
			if got := f.ResourceName(); got != tt.want {
				t.Errorf("FlowRule.ResourceName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlowRule_String(t *testing.T) {
	type fields struct {
		ID                uint64
		Resource          string
		LimitOrigin       string
		MetricType        MetricType
		Count             float64
		RelationStrategy  RelationStrategy
		ControlBehavior   ControlBehavior
		RefResource       string
		WarmUpPeriodSec   uint32
		MaxQueueingTimeMs uint32
		ClusterMode       bool
		ClusterConfig     ClusterRuleConfig
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			"test1",
			fields{0, "Sentinel", "", QPS, 0, Direct, Reject, "", 10, 1000, true, ClusterRuleConfig{AvgLocalThreshold}},
			`{"resource":"Sentinel","limitApp":"","grade":1,"count":0,"strategy":0,"controlBehavior":0,"warmUpPeriodSec":10,"maxQueueingTimeMs":1000,"clusterMode":true,"clusterConfig":{"thresholdType":0}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FlowRule{
				ID:                tt.fields.ID,
				Resource:          tt.fields.Resource,
				LimitOrigin:       tt.fields.LimitOrigin,
				MetricType:        tt.fields.MetricType,
				Count:             tt.fields.Count,
				RelationStrategy:  tt.fields.RelationStrategy,
				ControlBehavior:   tt.fields.ControlBehavior,
				RefResource:       tt.fields.RefResource,
				WarmUpPeriodSec:   tt.fields.WarmUpPeriodSec,
				MaxQueueingTimeMs: tt.fields.MaxQueueingTimeMs,
				ClusterMode:       tt.fields.ClusterMode,
				ClusterConfig:     tt.fields.ClusterConfig,
			}
			if got := f.String(); got != tt.want {
				t.Errorf("FlowRule.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

//Initialize FlowRule from JSON string
// After deserializing from JSON, determine the validity of each field
func TestFlowRule_LoadFromJson(t *testing.T) {
	// Resource test
	{
		rule := []byte(`{"resource":"Sentinel","limitApp":"","grade":1,"count":0,"strategy":0,"controlBehavior":0,"warmUpPeriodSec":10,"maxQueueingTimeMs":1000,"clusterMode":true,"clusterConfig":{"thresholdType":1}}`)
		flowRule := &FlowRule{}
		if err := json.Unmarshal(rule, &flowRule); err != nil {
			t.Errorf("load json error:%v", err)
		} else {
			if flowRule.ResourceName() != "Sentinel" {
				t.Error("unmarhal json ResourceName error")
			}
		}
	}
	// MetricType test
	{
		rule := []byte(`{"resource":"Sentinel","limitApp":"","grade":1,"count":0,"strategy":0,"controlBehavior":0,"warmUpPeriodSec":10,"maxQueueingTimeMs":1000,"clusterMode":true,"clusterConfig":{"thresholdType":1}}`)
		flowRule := &FlowRule{}
		if err := json.Unmarshal(rule, &flowRule); err != nil {
			t.Errorf("load json error:%v", err)
		} else {
			if flowRule.MetricType != Concurrency && flowRule.MetricType != QPS {
				t.Error("unmarhal json MetricType error，values: Concurrency(0) or QPS(1)")
			}
		}
	}
	// other property test ....
}
