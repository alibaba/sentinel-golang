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

package wtinylfu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoorkeeper(t *testing.T) {
	t.Run("Test_Doorkeeper", func(t *testing.T) {
		max := 1500
		filter := newDoorkeeper(1500, 0.001)
		for i := 0; i < max; i++ {
			filter.put(uint64(i))
			assert.True(t, true == filter.contains(uint64(i)))
		}
		filter.reset()
		for i := 0; i < max; i++ {
			assert.True(t, false == filter.contains(uint64(i)))
		}
	})
}
