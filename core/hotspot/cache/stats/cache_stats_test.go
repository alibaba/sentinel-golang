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

package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCacheStats(t *testing.T) {
	t.Run("Test_CacheStats", func(t *testing.T) {
		cs := NewCacheStats()
		for i := 0; i < 100; i++ {
			cs.RecordMisses()
		}
		assert.True(t, cs.MissCount() == 100)
		assert.True(t, cs.MissRate() == 1.0)
		assert.True(t, cs.HitRate() == 0.0)

		for i := 0; i < 100; i++ {
			cs.RecordHits()
		}

		assert.True(t, cs.MissCount() == 100)
		assert.True(t, cs.HitCount() == 100)
		assert.True(t, cs.MissRate() == 0.5)
		assert.True(t, cs.HitRate() == 0.5)

		for i := 0; i < 50; i++ {
			cs.RecordEviction()
		}
		assert.True(t, cs.EvictionCount() == 50)

		csNap := cs.Snapshot()
		for i := 0; i < 50; i++ {
			cs.RecordEviction()
		}
		assert.True(t, cs.EvictionCount() == 100)
		assert.True(t, csNap.EvictionCount() == 50)
	})
}
