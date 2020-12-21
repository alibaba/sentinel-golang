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
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/stretchr/testify/mock"
)

type TrafficShapingControllerMock struct {
	mock.Mock
}

func (m *TrafficShapingControllerMock) PerformChecking(arg interface{}, batchCount int64) *base.TokenResult {
	retArgs := m.Called(arg, batchCount)
	return retArgs.Get(0).(*base.TokenResult)
}

func (m *TrafficShapingControllerMock) BoundParamIndex() int {
	retArgs := m.Called()
	return retArgs.Int(0)
}

func (m *TrafficShapingControllerMock) BoundMetric() *ParamsMetric {
	retArgs := m.Called()
	return retArgs.Get(0).(*ParamsMetric)
}

func (m *TrafficShapingControllerMock) BoundRule() *Rule {
	retArgs := m.Called()
	return retArgs.Get(0).(*Rule)
}

func (m *TrafficShapingControllerMock) Replace(r *Rule) {
	_ = m.Called(r)
	return
}

func (m *TrafficShapingControllerMock) ExtractArgs(ctx *base.EntryContext) []interface{} {
	_ = m.Called()
	ret := []interface{}{ctx.Input.Args[m.BoundParamIndex()]}
	return ret
}
