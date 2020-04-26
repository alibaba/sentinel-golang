package datasource

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestFlowRulesJsonConverter(t *testing.T) {
	// Prepare test data
	f, err := os.Open("../../tests/testdata/extension/plugin/FlowRule.json")
	defer f.Close()
	if err != nil {
		t.Errorf("The rules file is not existed, err:%+v.", errors.WithStack(err))
	}
	normalSrc, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("Fail to read file, err: %+v.", errors.WithStack(err))
	}
	normalWant := make([]*flow.FlowRule, 0)
	err = json.Unmarshal(normalSrc, &normalWant)
	if err != nil {
		t.Errorf("Fail to unmarshal source:%+v to []flow.FlowRule, err:%+v", normalSrc, err)
	}

	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "TestFlowRulesJsonConverter_Nil",
			args: args{
				src: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "TestFlowRulesJsonConverter_Unmarshal_Error",
			args: args{
				src: []byte("{1111111}"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "TestFlowRulesJsonConverter_Normal",
			args: args{
				src: normalSrc,
			},
			want:    normalWant,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FlowRulesJsonConverter(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("FlowRulesJsonConverter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FlowRulesJsonConverter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlowRulesUpdater(t *testing.T) {
	t.Run("TestFlowRulesUpdater_Nil", func(t *testing.T) {
		flow.ClearRules()
		flow.LoadRules([]*flow.FlowRule{
			{
				ID:                0,
				Resource:          "abc",
				LimitOrigin:       "default",
				MetricType:        0,
				Count:             0,
				RelationStrategy:  0,
				ControlBehavior:   0,
				RefResource:       "",
				WarmUpPeriodSec:   0,
				MaxQueueingTimeMs: 0,
				ClusterMode:       false,
				ClusterConfig:     flow.ClusterRuleConfig{},
			}})
		assert.True(t, len(flow.GetRules()) == 1, "Fail to prepare test data.")
		err := FlowRulesUpdater(nil)
		assert.True(t, err == nil && len(flow.GetRules()) == 0, "Fail to test TestFlowRulesUpdater_Nil")
	})

	t.Run("TestFlowRulesUpdater_Assert_Failed", func(t *testing.T) {
		flow.ClearRules()
		err := FlowRulesUpdater("xxxxxxxx")
		assert.True(t, err != nil && strings.Contains(err.Error(), "Fail to type assert data to []flow.FlowRule"))
	})

	t.Run("TestFlowRulesUpdater_Empty_Rules", func(t *testing.T) {
		flow.ClearRules()
		p := make([]flow.FlowRule, 0)
		err := FlowRulesUpdater(p)
		assert.True(t, err == nil && len(flow.GetRules()) == 0)
	})

	t.Run("TestFlowRulesUpdater_Normal", func(t *testing.T) {
		flow.ClearRules()
		p := make([]flow.FlowRule, 0)
		fw := flow.FlowRule{
			ID:                0,
			Resource:          "aaaa",
			LimitOrigin:       "aaa",
			MetricType:        0,
			Count:             0,
			RelationStrategy:  0,
			ControlBehavior:   0,
			RefResource:       "",
			WarmUpPeriodSec:   0,
			MaxQueueingTimeMs: 0,
			ClusterMode:       false,
			ClusterConfig:     flow.ClusterRuleConfig{},
		}
		p = append(p, fw)
		err := FlowRulesUpdater(p)
		assert.True(t, err == nil && len(flow.GetRules()) == 1)
	})
}

func TestSystemRulesJsonConvert(t *testing.T) {
	// Prepare test data
	f, err := os.Open("../../tests/testdata/extension/plugin/SystemRule.json")
	defer f.Close()
	if err != nil {
		t.Errorf("The rules file is not existed, err:%+v.", errors.WithStack(err))
	}
	normalSrc, err := ioutil.ReadAll(f)
	if err != nil {
		t.Errorf("Fail to read file, err: %+v.", errors.WithStack(err))
	}
	normalWant := make([]*system.SystemRule, 0)
	err = json.Unmarshal(normalSrc, &normalWant)
	if err != nil {
		t.Errorf("Fail to unmarshal source:%+v to []system.SystemRule, err:%+v", normalSrc, err)
	}

	type args struct {
		src []byte
	}
	tests := []struct {
		name    string
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "TestSystemRulesJsonConvert_Nil",
			args: args{
				src: nil,
			},
			want:    nil,
			wantErr: false,
		},
		{
			name: "TestSystemRulesJsonConvert_Unmarshal_Error",
			args: args{
				src: []byte("{1111111}"),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "TestSystemRulesJsonConvert_Normal",
			args: args{
				src: normalSrc,
			},
			want:    normalWant,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SystemRulesJsonConverter(tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("SystemRulesJsonConverter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SystemRulesJsonConverter() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSystemRulesUpdater(t *testing.T) {
	t.Run("TestSystemRulesUpdater_Nil", func(t *testing.T) {
		system.ClearRules()
		system.LoadRules([]*system.SystemRule{
			{
				ID:           0,
				MetricType:   0,
				TriggerCount: 0,
				Strategy:     0,
			},
		})
		assert.True(t, len(system.GetRules()) == 1, "Fail to prepare data.")
		err := SystemRulesUpdater(nil)
		assert.True(t, err == nil && len(system.GetRules()) == 0, "Fail to test TestSystemRulesUpdater_Nil")
	})

	t.Run("TestSystemRulesUpdater_Assert_Failed", func(t *testing.T) {
		system.ClearRules()
		err := SystemRulesUpdater("xxxxxxxx")
		assert.True(t, err != nil && strings.Contains(err.Error(), "Fail to type assert data to []system.SystemRule"))
	})

	t.Run("TestSystemRulesUpdater_Empty_Rules", func(t *testing.T) {
		system.ClearRules()
		p := make([]system.SystemRule, 0)
		err := SystemRulesUpdater(p)
		assert.True(t, err == nil && len(system.GetRules()) == 0)
	})

	t.Run("TestSystemRulesUpdater_Normal", func(t *testing.T) {
		system.ClearRules()
		p := make([]system.SystemRule, 0)
		sr := system.SystemRule{
			ID:           0,
			MetricType:   0,
			TriggerCount: 0,
			Strategy:     0,
		}
		p = append(p, sr)
		err := SystemRulesUpdater(p)
		assert.True(t, err == nil && len(system.GetRules()) == 1)
	})
}
