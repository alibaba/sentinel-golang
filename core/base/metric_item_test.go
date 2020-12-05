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

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetricItemFromFatStringLegal(t *testing.T) {
	line1 := "1564382218000|2019-07-29 14:36:58|/foo/*|4|9|3|0|25|0|2|1"
	item1, err := MetricItemFromFatString(line1)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1564382218000), item1.Timestamp)
	assert.Equal(t, uint64(4), item1.PassQps)
	assert.Equal(t, uint64(9), item1.BlockQps)
	assert.Equal(t, uint64(3), item1.CompleteQps)
	assert.Equal(t, uint64(0), item1.ErrorQps)
	assert.Equal(t, uint64(25), item1.AvgRt)
	assert.Equal(t, "/foo/*", item1.Resource)
	assert.Equal(t, int32(1), item1.Classification)
}

func TestMetricItemFromFatStringIllegal(t *testing.T) {
	line1 := "1564382218000|2019-07-29 14:36:58|foo|baz|4|9|3|0|25|0|2|1"
	_, err := MetricItemFromFatString(line1)
	assert.Error(t, err, "Error should occur when parsing malformed line")

	line2 := "1564382218000|2019-07-29 14:36:58|foo|-3|9|3|0|25|0|2|1"
	_, err = MetricItemFromFatString(line2)
	assert.Error(t, err, "Error should occur when parsing malformed line")
}
