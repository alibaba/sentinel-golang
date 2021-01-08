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

func TestCountMinSketch(t *testing.T) {
	t.Run("Test_CountMinSketch", func(t *testing.T) {
		max := 15
		cm4 := newCountMinSketch(max)
		for i := 0; i < max; i++ {
			for j := i; j > 0; j-- {
				cm4.add(uint64(i))
			}
			assert.True(t, uint64(i) == uint64(cm4.estimate(uint64(i))))
		}

		cm4.reset()
		for i := 0; i < max; i++ {
			assert.True(t, uint64(i)/2 == uint64(cm4.estimate(uint64(i))))
		}

		cm4.clear()
		for i := 0; i < max; i++ {
			assert.True(t, 0 == uint64(cm4.estimate(uint64(i))))
		}
	})
}
