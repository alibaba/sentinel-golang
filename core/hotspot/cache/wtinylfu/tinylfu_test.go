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

type tinyLFUTest struct {
	lfu *TinyLfu
	t   *testing.T
}

func (t *tinyLFUTest) assertCap(n int) {
	assert.True(t.t, t.lfu.lru.cap+t.lfu.slru.protectedCap+t.lfu.slru.probationCap == n)
}

func (t *tinyLFUTest) assertLen(admission, protected, probation int) {
	sz := t.lfu.Len()
	tz := t.lfu.slru.protectedLs.Len()
	bz := t.lfu.slru.probationLs.Len()
	assert.True(t.t, sz == admission+protected+probation && tz == protected && bz == probation)
}

func (t *tinyLFUTest) assertLRUValue(k int, id uint8) {
	v := t.lfu.items[k].Value.(*slruItem).value
	assert.True(t.t, v != nil)
	ak := k
	av := v
	listId := t.lfu.items[k].Value.(*slruItem).listId
	assert.True(t.t, ak == av && listId == id)
}

func TestTinyLFU(t *testing.T) {
	t.Run("Test_TinyLFU", func(t *testing.T) {

		s := tinyLFUTest{t: t}
		s.lfu, _ = NewTinyLfu(200, false)
		s.assertCap(200)
		s.lfu.slru.protectedCap = 2
		s.lfu.slru.probationCap = 1
		for i := 0; i < 5; i++ {
			e := s.lfu.AddIfAbsent(i, i)
			assert.True(t, e == nil)
		}
		// 4 3 | - | 2 1 0
		s.assertLen(2, 0, 3)
		s.assertLRUValue(4, admissionWindow)
		s.assertLRUValue(3, admissionWindow)
		s.assertLRUValue(2, probationSegment)
		s.assertLRUValue(1, probationSegment)
		s.assertLRUValue(0, probationSegment)

		s.lfu.Get(1)
		s.lfu.Get(2)
		// 4 3 | 2 1 | 0
		s.assertLen(2, 2, 1)
		s.assertLRUValue(2, protectedSegment)
		s.assertLRUValue(1, protectedSegment)
		s.assertLRUValue(0, probationSegment)

		s.lfu.AddIfAbsent(5, 5)
		// 5 4 | 2 1 | 0
		s.assertLRUValue(5, admissionWindow)
		s.assertLRUValue(4, admissionWindow)
		s.assertLRUValue(2, protectedSegment)
		s.assertLRUValue(1, protectedSegment)
		s.assertLRUValue(0, probationSegment)

		s.lfu.Get(4)
		s.lfu.Get(5)
		s.lfu.AddIfAbsent(6, 6)
		// 6 5 | 2 1 | 4
		s.assertLRUValue(6, admissionWindow)
		s.assertLRUValue(5, admissionWindow)
		s.assertLRUValue(2, protectedSegment)
		s.assertLRUValue(1, protectedSegment)
		s.assertLRUValue(4, probationSegment)
		s.assertLen(2, 2, 1)
		n := s.lfu.estimate(sum(1))
		assert.True(t, n == 2)
		s.lfu.Get(2)
		s.lfu.Get(2)
		n = s.lfu.estimate(sum(2))
		assert.True(t, n == 4)
	})
}
