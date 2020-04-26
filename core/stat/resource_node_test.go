package stat

import (
	"fmt"
	"sync"
	"testing"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/assert"
)

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
