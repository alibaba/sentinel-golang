// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package flow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuleNeedStatistic(t *testing.T) {
	// need
	r1 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	// no need
	r2 := &Rule{
		Resource:               "abc1",
		Threshold:              200,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Throttling,
		MaxQueueingTimeMs:      10,
		StatIntervalInMs:       2000,
	}
	// need
	r3 := &Rule{
		Resource:               "abc1",
		Threshold:              300,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: WarmUp,
		ControlBehavior:        Reject,
		MaxQueueingTimeMs:      10,
		StatIntervalInMs:       5000,
	}
	// need
	r4 := &Rule{
		Resource:               "abc1",
		Threshold:              400,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: WarmUp,
		ControlBehavior:        Throttling,
		MaxQueueingTimeMs:      10,
		StatIntervalInMs:       50000,
	}

	assert.True(t, r1.needStatistic())
	assert.False(t, r2.needStatistic())
	assert.True(t, r3.needStatistic())
	assert.True(t, r4.needStatistic())
}

func TestRuleIsStatReusable(t *testing.T) {
	// Not same resource
	r11 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	r12 := &Rule{
		Resource:               "abc2",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	assert.False(t, r11.isStatReusable(r12))

	// Not same relation strategy
	r21 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	r22 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       AssociatedResource,
		RefResource:            "abc3",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	assert.False(t, r21.isStatReusable(r22))

	// Not same ref resource
	r31 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       AssociatedResource,
		RefResource:            "abc3",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	r32 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       AssociatedResource,
		RefResource:            "abc4",
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	assert.False(t, r31.isStatReusable(r32))

	// Not same stat interval
	r41 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	r42 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       2000,
	}
	assert.False(t, r41.isStatReusable(r42))

	// Not both need stat
	r51 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	r52 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Throttling,
		StatIntervalInMs:       1000,
	}
	assert.False(t, r51.isStatReusable(r52))

	// Not same threshold
	r61 := &Rule{
		Resource:               "abc1",
		Threshold:              100,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	r62 := &Rule{
		Resource:               "abc1",
		Threshold:              200,
		RelationStrategy:       CurrentResource,
		TokenCalculateStrategy: Direct,
		ControlBehavior:        Reject,
		StatIntervalInMs:       1000,
	}
	assert.True(t, r61.isStatReusable(r62))
}
