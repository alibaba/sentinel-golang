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

package cache

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_concurrentLruCounterCacheMap_Add_Get(t *testing.T) {
	t.Run("Test_concurrentLruCounterCacheMap_Add_Get", func(t *testing.T) {
		c := NewLRUCacheMap(100)
		for i := 1; i <= 100; i++ {
			val := int64(i)
			c.Add(strconv.Itoa(i), &val)
		}
		assert.True(t, c.Len() == 100)
		val, found := c.Get("1")
		assert.True(t, found && *val == 1)
	})
}

func Test_concurrentLruCounterCacheMap_AddIfAbsent(t *testing.T) {
	t.Run("Test_concurrentLruCounterCacheMap_AddIfAbsent", func(t *testing.T) {
		c := NewLRUCacheMap(100)
		for i := 1; i <= 99; i++ {
			val := int64(i)
			c.Add(strconv.Itoa(i), &val)
		}
		v := int64(100)
		prior := c.AddIfAbsent("100", &v)
		assert.True(t, prior == nil)
		v2 := int64(1000)
		prior2 := c.AddIfAbsent("100", &v2)
		assert.True(t, *prior2 == 100)
	})
}

func Test_concurrentLruCounterCacheMap_Contains(t *testing.T) {
	t.Run("Test_concurrentLruCounterCacheMap_Contains", func(t *testing.T) {
		c := NewLRUCacheMap(100)
		for i := 1; i <= 100; i++ {
			val := int64(i)
			c.Add(strconv.Itoa(i), &val)
		}
		assert.True(t, c.Contains("100") == true)
		assert.True(t, c.Contains("1") == true)
		assert.True(t, c.Contains("101") == false)

		val := int64(101)
		c.Add(strconv.Itoa(int(val)), &val)
		assert.True(t, c.Contains("1") == false)
	})
}

func Test_concurrentLruCounterCacheMap_Keys(t *testing.T) {
	t.Run("Test_concurrentLruCounterCacheMap_Add", func(t *testing.T) {
		c := NewLRUCacheMap(100)
		for i := 1; i <= 100; i++ {
			val := int64(i)
			c.Add(strconv.Itoa(i), &val)
		}
		assert.True(t, len(c.Keys()) == 100)
		assert.True(t, c.Keys()[0] == "1")
		assert.True(t, c.Keys()[99] == "100")
	})
}

func Test_concurrentLruCounterCacheMap_Purge(t *testing.T) {
	t.Run("Test_concurrentLruCounterCacheMap_Add", func(t *testing.T) {
		c := NewLRUCacheMap(100)
		for i := 1; i <= 100; i++ {
			val := int64(i)
			c.Add(strconv.Itoa(i), &val)
		}
		assert.True(t, c.Len() == 100)
		c.Purge()
		assert.True(t, c.Len() == 0)
	})
}

func Test_concurrentLruCounterCacheMap_Remove(t *testing.T) {
	t.Run("Test_concurrentLruCounterCacheMap_Add", func(t *testing.T) {
		c := NewLRUCacheMap(100)
		for i := 1; i <= 100; i++ {
			val := int64(i)
			c.Add(strconv.Itoa(i), &val)
		}
		assert.True(t, c.Len() == 100)

		c.Remove("100")
		assert.True(t, c.Len() == 99)
		val, existed := c.Get("100")
		assert.True(t, existed == false && val == nil)
	})
}
