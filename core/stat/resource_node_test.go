package stat

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	sbase "github.com/alibaba/sentinel-golang/core/stat/base"

	"github.com/stretchr/testify/assert"
)

func TestNewResourceNode(t *testing.T) {
	rn := NewResourceNode("test", base.ResTypeCommon)
	assert.Equal(t, "test", rn.resourceName)
	assert.Equal(t, base.ResTypeCommon, rn.resourceType)
	assert.NotNil(t, rn.BaseStatNode)
	assert.NotNil(t, rn.readOnlyStats)
	assert.Equal(t, 0, len(rn.readOnlyStats))
}

func TestResourceNodeResourceType(t *testing.T) {
	rn := &ResourceNode{
		resourceType: base.ResTypeCommon,
	}
	assert.Equal(t, base.ResTypeCommon, rn.ResourceType())
}

func TestResourceNodeResourceName(t *testing.T) {
	rn := &ResourceNode{
		resourceName: "test",
	}
	assert.Equal(t, "test", rn.ResourceName())
}

func TestResourceNodeGetSlidingWindowMetric(t *testing.T) {
	s := &sbase.SlidingWindowMetric{}
	tests := []struct {
		name          string
		key           string
		readOnlyStats map[string]*sbase.SlidingWindowMetric
		expected      *sbase.SlidingWindowMetric
	}{
		{
			name:          "EmptyKeyReadOnlyStats",
			key:           "test1",
			readOnlyStats: make(map[string]*sbase.SlidingWindowMetric),
			expected:      nil,
		},
		{
			name: "EmptyValueInReadOnlyStats",
			key:  "test1",
			readOnlyStats: map[string]*sbase.SlidingWindowMetric{
				"test1": nil,
			},
			expected: nil,
		},
		{
			name: "NormalValueInReadOnlyStats",
			key:  "test1",
			readOnlyStats: map[string]*sbase.SlidingWindowMetric{
				"test1": s,
			},
			expected: s,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rn := &ResourceNode{
				readOnlyStats: test.readOnlyStats,
			}
			assert.Equal(t, test.expected, rn.GetSlidingWindowMetric(test.key))
		})
	}
}

func TestResourceNodeGetOrCreateSlidingWindowMetric(t *testing.T) {
	rn := NewResourceNode("test", base.ResTypeCommon)
	swm := sbase.NewSlidingWindowMetric(2, 2000, rn.arr)
	rn.readOnlyStats = map[string]*sbase.SlidingWindowMetric{
		"102000": swm,
	}
	swm1 := sbase.NewSlidingWindowMetric(2, 10000, rn.arr)
	tests := []struct {
		name         string
		sampleCount  uint32
		intervalInMs uint32
		expected     *sbase.SlidingWindowMetric
	}{
		{
			name:         "ExistKey",
			sampleCount:  2,
			intervalInMs: 2000,
			expected:     swm,
		},
		{
			name:         "NonExistKey",
			sampleCount:  2,
			intervalInMs: 10000,
			expected:     swm1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expected, rn.GetOrCreateSlidingWindowMetric(test.sampleCount, test.intervalInMs))
			assert.Equal(t, test.expected, rn.GetSlidingWindowMetric(fmt.Sprintf("%d/%d", test.sampleCount, test.intervalInMs)))
		})
	}
}

func TestResourceNode_GetOrCreateSlidingWindowMetric(t *testing.T) {
	type args struct {
		sampleCount  uint32
		intervalInMs uint32
	}
	tests := []struct {
		name string
	}{
		{
			name: "TestResourceNode_GetOrCreateSlidingWindowMetric",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewResourceNode("aa", base.ResTypeCommon)

			argsList := []args{
				{
					sampleCount:  10,
					intervalInMs: 10000,
				},
				{
					sampleCount:  5,
					intervalInMs: 10000,
				},
				{
					sampleCount:  2,
					intervalInMs: 10000,
				},
				{
					sampleCount:  1,
					intervalInMs: 10000,
				},
				{
					sampleCount:  10,
					intervalInMs: 5000,
				},
				{
					sampleCount:  2,
					intervalInMs: 5000,
				},
				{
					sampleCount:  5,
					intervalInMs: 5000,
				},
				{
					sampleCount:  1,
					intervalInMs: 5000,
				},
				{
					sampleCount:  1,
					intervalInMs: 2000,
				},
				{
					sampleCount:  2,
					intervalInMs: 2000,
				},
			}

			wg := &sync.WaitGroup{}
			wg.Add(100)
			for i := 0; i < 100; i++ {
				go func(g *sync.WaitGroup) {
					for _, as := range argsList {
						n.GetOrCreateSlidingWindowMetric(as.sampleCount, as.intervalInMs)
					}
					g.Done()
				}(wg)
			}
			wg.Wait()

			for _, as := range argsList {
				key := fmt.Sprintf("%d/%d", as.sampleCount, as.intervalInMs)
				assert.True(t, n.GetSlidingWindowMetric(key) != nil)
			}
			assert.True(t, len(n.readOnlyStats) == 10)
		})
	}
}
