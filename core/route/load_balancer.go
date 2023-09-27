package route

import (
	"github.com/alibaba/sentinel-golang/core/route/base"
	"math/rand"
	"sync"
)

type LoadBalancer interface {
	Select(instances []*base.Instance, context *base.TrafficContext) (*base.Instance, error)
}

type RandomLoadBalancer struct {
}

func NewRandomLoadBalancer() *RandomLoadBalancer {
	return &RandomLoadBalancer{}
}

func (r *RandomLoadBalancer) Select(instances []*base.Instance, context *base.TrafficContext) (*base.Instance, error) {
	if len(instances) == 0 {
		return nil, nil
	}

	return instances[rand.Intn(len(instances))], nil
}

type RoundRobinLoadBalancer struct {
	idx int
	mu  sync.Mutex
}

func NewRoundRobinLoadBalancer() *RoundRobinLoadBalancer {
	return &RoundRobinLoadBalancer{idx: 0}
}

func (r *RoundRobinLoadBalancer) Select(instances []*base.Instance, context *base.TrafficContext) (*base.Instance, error) {
	if len(instances) == 0 {
		return nil, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.idx = (r.idx + 1) % len(instances)
	return instances[r.idx], nil
}
