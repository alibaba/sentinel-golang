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

package system_metric

import (
	"sync"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

func TestCurrentLoad(t *testing.T) {
	defer currentLoad.Store(NotRetrievedLoadValue)

	cLoad := CurrentLoad()
	assert.True(t, util.Float64Equals(NotRetrievedLoadValue, cLoad))

	v := float64(1.0)
	currentLoad.Store(v)
	cLoad = CurrentLoad()
	assert.True(t, util.Float64Equals(v, cLoad))
}

func TestCurrentCpuUsage(t *testing.T) {
	defer currentCpuUsage.Store(NotRetrievedCpuUsageValue)

	cpuUsage := CurrentCpuUsage()
	assert.Equal(t, NotRetrievedCpuUsageValue, cpuUsage)

	v := float64(0.3)
	currentCpuUsage.Store(v)
	cpuUsage = CurrentCpuUsage()
	assert.True(t, util.Float64Equals(v, cpuUsage))
}

func Test_getProcessCpuStat(t *testing.T) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		i := 0
		wg.Done()
		for i < 10000000000 {
			i++
			if i == 1000000000 {
				i = 0
			}
		}
	}()
	wg.Wait()

	got, err := getProcessCpuStat()
	if err != nil {
		t.Error(err)
	}
	assert.True(t, int(got) == 0)
	time.Sleep(time.Millisecond * 200)

	got, err = getProcessCpuStat()
	if err != nil {
		t.Error(err)
	}
	assert.True(t, int(got) > 0)
	time.Sleep(time.Millisecond * 200)

	got, err = getProcessCpuStat()
	if err != nil {
		t.Error(err)
	}
	assert.True(t, int(got) > 0)
	time.Sleep(time.Millisecond * 200)
}
