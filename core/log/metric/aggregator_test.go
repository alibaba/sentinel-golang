package metric

import (
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/stat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	defaultTestResourceName = "abc"
)

func Test_aggregateIntoMap(t *testing.T) {
	type args struct {
		mm      metricTimeMap
		metrics map[uint64]*base.MetricItem
		node    *stat.ResourceNode
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "Test_aggregateIntoMap",
			args: args{
				mm:      make(metricTimeMap),
				metrics: make(map[uint64]*base.MetricItem),
				node:    stat.NewResourceNode(defaultTestResourceName, base.ResTypeCommon),
			},
		},
	}
	mi1 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  0,
		Timestamp:       1581959010000,
		PassQps:         10,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	mi2 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  0,
		Timestamp:       1581959011000,
		PassQps:         20,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	mi3 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  0,
		Timestamp:       1581959012000,
		PassQps:         30,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	mi4 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  1,
		Timestamp:       1581959012000,
		PassQps:         60,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	tests[0].args.metrics[mi1.Timestamp] = mi1
	tests[0].args.metrics[mi2.Timestamp] = mi2
	tests[0].args.metrics[mi3.Timestamp] = mi3
	tests[0].args.mm[mi4.Timestamp] = []*base.MetricItem{mi4}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aggregateIntoMap(tt.args.mm, tt.args.metrics, tt.args.node)
			assert.True(t, len(tt.args.mm[mi1.Timestamp]) == 1)
			assert.True(t, len(tt.args.mm[mi2.Timestamp]) == 1)
			assert.True(t, len(tt.args.mm[mi3.Timestamp]) == 2)
		})
	}
}

func Test_isActiveMetricItem(t *testing.T) {
	type args struct {
		item *base.MetricItem
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_isActiveMetricItem",
			args: args{
				item: &base.MetricItem{
					Resource:        "abc",
					Classification:  0,
					Timestamp:       0,
					PassQps:         1,
					BlockQps:        0,
					CompleteQps:     0,
					ErrorQps:        0,
					AvgRt:           0,
					OccupiedPassQps: 0,
					Concurrency:     0,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isActiveMetricItem(tt.args.item); got != tt.want {
				t.Errorf("isActiveMetricItem() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isItemTimestampInTime(t *testing.T) {
	type args struct {
		ts              uint64
		currentSecStart uint64
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Test_isItemTimestampInTime_true",
			args: args{
				ts:              1581959014000,
				currentSecStart: 1581959015000,
			},
			want: true,
		},
		{
			name: "Test_isItemTimestampInTime_false",
			args: args{
				ts:              1581959014000,
				currentSecStart: 1581959014000,
			},
			want: false,
		},
	}
	lastFetchTime = 1581959013000
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isItemTimestampInTime(tt.args.ts, tt.args.currentSecStart); got != tt.want {
				t.Errorf("isItemTimestampInTime() = %v, want %v", got, tt.want)
			}
		})
	}
}

type MetricItemRetrieverMock struct {
	mock.Mock
}

func (m *MetricItemRetrieverMock) MetricsOnCondition(predicate base.TimePredicate) []*base.MetricItem {
	args := m.Called(predicate)
	return args.Get(0).([]*base.MetricItem)
}

func Test_currentMetricItems(t *testing.T) {
	type args struct {
		retriever   base.MetricItemRetriever
		currentTime uint64
	}
	tests := []struct {
		name string
		args args
		want map[uint64]*base.MetricItem
	}{
		{
			name: "Test_currentMetricItems",
			args: args{
				retriever:   nil,
				currentTime: 1581959014000,
			},
			want: nil,
		},
	}

	m := &MetricItemRetrieverMock{}
	tests[0].args.retriever = m
	ret := make([]*base.MetricItem, 0, 2)
	mi1 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  0,
		Timestamp:       1581959010000,
		PassQps:         10,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	mi2 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  0,
		Timestamp:       1581959011000,
		PassQps:         20,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	mi3 := &base.MetricItem{
		Resource:        defaultTestResourceName,
		Classification:  0,
		Timestamp:       1581959012000,
		PassQps:         0,
		BlockQps:        0,
		CompleteQps:     0,
		ErrorQps:        0,
		AvgRt:           0,
		OccupiedPassQps: 0,
		Concurrency:     0,
	}
	ret = append(ret, mi1, mi2, mi3)
	m.On("MetricsOnCondition", mock.Anything).Return(ret)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := currentMetricItems(tt.args.retriever, tt.args.currentTime)
			if len(got) != 2 {
				t.Errorf("get map len = %v, want %v", len(got), 2)
			}
			if got[1581959010000] == nil || got[1581959011000] == nil {
				t.Errorf("result error, %v", got)
			}
		})
	}
}
