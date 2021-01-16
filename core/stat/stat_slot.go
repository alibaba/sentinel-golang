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

package stat

import (
	"github.com/alibaba/sentinel-golang/core/base"
	metric_exporter "github.com/alibaba/sentinel-golang/exporter/metric"
	"github.com/alibaba/sentinel-golang/util"
)

const (
	StatSlotName  = "sentinel-core-stat-slot"
	StatSlotOrder = 1000
)

var (
	DefaultSlot = &Slot{}

	passCounter = metric_exporter.NewCounter(
		"pass",
		"Total pass count",
		[]string{"resource"})
	blockCounter = metric_exporter.NewCounter(
		"block",
		"Total block count",
		[]string{"resource", "block_type"})
	completeCounter = metric_exporter.NewCounter(
		"complete",
		"Total complete count",
		[]string{"resource"})
	errorCounter = metric_exporter.NewCounter(
		"error",
		"Total error count",
		[]string{"resource"})
	rtHistogram = metric_exporter.NewHistogram(
		"rt",
		"Rt histogram",
		[]float64{1.0, 5.0, 10.0, 50.0, 100.0, 500.0, 1000, 5000},
		[]string{"resource"})
)

func init() {
	metric_exporter.MustRegister(passCounter)
	metric_exporter.MustRegister(blockCounter)
	metric_exporter.MustRegister(completeCounter)
	metric_exporter.MustRegister(errorCounter)
	metric_exporter.MustRegister(rtHistogram)
}

type Slot struct {
}

func (s *Slot) Name() string {
	return StatSlotName
}

func (s *Slot) Order() uint32 {
	return StatSlotOrder
}

func (s *Slot) OnEntryPassed(ctx *base.EntryContext) {
	s.recordPassFor(ctx.StatNode, ctx.Input.BatchCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordPassFor(InboundNode(), ctx.Input.BatchCount)
	}

	passCounter.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name())
}

func (s *Slot) OnEntryBlocked(ctx *base.EntryContext, blockError *base.BlockError) {
	s.recordBlockFor(ctx.StatNode, ctx.Input.BatchCount)
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordBlockFor(InboundNode(), ctx.Input.BatchCount)
	}

	blockCounter.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name(), blockError.BlockType().String())
}

func (s *Slot) OnCompleted(ctx *base.EntryContext) {
	rt := util.CurrentTimeMillis() - ctx.StartTime()
	ctx.PutRt(rt)
	s.recordCompleteFor(ctx.StatNode, ctx.Input.BatchCount, rt, ctx.Err())
	if ctx.Resource.FlowType() == base.Inbound {
		s.recordCompleteFor(InboundNode(), ctx.Input.BatchCount, rt, ctx.Err())
	}

	completeCounter.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name())
	if ctx.Err() != nil {
		errorCounter.Add(float64(ctx.Input.BatchCount), ctx.Resource.Name())
	}
	rtHistogram.Observe(float64(rt), ctx.Resource.Name())
}

func (s *Slot) recordPassFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.IncreaseConcurrency()
	sn.AddCount(base.MetricEventPass, int64(count))
}

func (s *Slot) recordBlockFor(sn base.StatNode, count uint32) {
	if sn == nil {
		return
	}
	sn.AddCount(base.MetricEventBlock, int64(count))
}

func (s *Slot) recordCompleteFor(sn base.StatNode, count uint32, rt uint64, err error) {
	if sn == nil {
		return
	}
	if err != nil {
		sn.AddCount(base.MetricEventError, int64(count))
	}
	sn.AddCount(base.MetricEventRt, int64(rt))
	sn.AddCount(base.MetricEventComplete, int64(count))
	sn.DecreaseConcurrency()
}
