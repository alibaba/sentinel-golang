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

package adaptive

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegistryCalculateStrategy(t *testing.T) {
	calculateStrategyNew1 := CalculateStrategy(600)
	err := RegistryCalculateStrategy(calculateStrategyNew1, "test")
	assert.Nil(t, err)
	assert.True(t, calculateStrategyNew1.String() == "test")

	MetricTypeNew1 := MetricType(600)
	err = RegistryMetricType(MetricTypeNew1, "test")
	assert.Nil(t, err)
	assert.True(t, MetricTypeNew1.String() == "test")
}
