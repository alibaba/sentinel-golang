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

package hotspot

import (
	"time"

	"github.com/alibaba/sentinel-golang/core/stat"

	"github.com/alibaba/sentinel-golang/core/base"
)

const (
	RuleCheckSlotName  = "sentinel-core-hotspot-rule-check-slot"
	RuleCheckSlotOrder = 4000
)

var (
	DefaultSlot = &Slot{}
)

type Slot struct {
}

func (s *Slot) Name() string {
	return RuleCheckSlotName
}

func (s *Slot) Order() uint32 {
	return RuleCheckSlotOrder
}

func (s *Slot) Check(ctx *base.EntryContext) *base.TokenResult {
	res := ctx.Resource.Name()
	batch := int64(ctx.Input.BatchCount)
	result := ctx.RuleCheckResult
	nodes := stat.GetAllNodes(ctx)
	for _, node := range nodes {
		tcs := getTrafficControllersFor(node.ResourceName())
		for _, tc := range tcs {
			args := tc.ExtractArgs(ctx)
			if args == nil || len(args) == 0 {
				continue
			}

			for _, arg := range args {
				arg := arg
				r := canPassCheck(tc, arg, batch)
				if r == nil {
					continue
				}

				if r.Status() == base.ResultStatusBlocked {
					if tc.BoundRule().Mode == MONITOR {
						stat.SetFirstMonitorNodeAndRule(ctx, node, tc.BoundRule())
						continue
					}
					stat.SetBlockNode(ctx, node)
					r.ResetToBlockedWith(
						base.WithBlockResource(res),
						base.WithBlockType(base.BlockTypeHotSpotParamFlow),
						base.WithRule(tc.BoundRule()),
						base.WithBlockResource(node.ResourceName()))
					return r

				}
				if r.Status() == base.ResultStatusShouldWait {
					if nanosToWait := r.NanosToWait(); nanosToWait > 0 {
						// Handle waiting action.
						time.Sleep(nanosToWait)
					}
					continue
				}
			}
		}
	}
	return result
}

func canPassCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return canPassLocalCheck(tc, arg, batch)
}

func canPassLocalCheck(tc TrafficShapingController, arg interface{}, batch int64) *base.TokenResult {
	return tc.PerformChecking(arg, batch)
}
