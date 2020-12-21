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
	"fmt"
	"math/rand"
	"testing"

	lru2 "github.com/alibaba/sentinel-golang/core/hotspot/cache/lru"
)

func testBySize(cacheSize int, zipf *rand.Zipf) {
	lfu, _ := NewTinyLfu(cacheSize, true)
	lru, _ := lru2.NewLRU(cacheSize, nil, true)
	for i := 0; i < 2000000; i++ {
		key := zipf.Uint64()
		_, ok := lfu.Get(key)
		if !ok {
			lfu.Add(key, key)
		}
	}
	s, _ := lfu.Stats()

	fmt.Printf("tinyLfu cache size %d, hit %d, miss %d, evictionCount %d, hitRate %f \n", cacheSize, s.HitCount(),
		s.MissCount(), s.EvictionCount(), s.HitRate())

	for i := 0; i < 2000000; i++ {
		key := zipf.Uint64()
		_, ok := lru.Get(key)
		if !ok {
			lru.Add(key, key)
		}
	}
	st, _ := lru.Stats()
	fmt.Printf("lru cache size %d, hit %d, miss %d, evictionCount %d, hitRate %f \n", cacheSize, st.HitCount(),
		st.MissCount(), st.EvictionCount(), st.HitRate())
}

func TestHitRate(t *testing.T) {
	t.Run("Test_HitRate", func(t *testing.T) {
		r := rand.New(rand.NewSource(1))
		zipf := rand.NewZipf(
			r,
			1.01,
			1.0,
			1<<18-1,
		)
		testBySize(100, zipf)
		testBySize(500, zipf)
		testBySize(1000, zipf)
		testBySize(5000, zipf)
		testBySize(10000, zipf)
		testBySize(20000, zipf)
		testBySize(50000, zipf)
	})
}
