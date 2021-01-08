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
	"sync/atomic"
)

// CacheStats is statistics about a cache.
type CacheStats struct {
	hitCount      *uint64
	missCount     *uint64
	evictionCount *uint64
}

func NewCacheStats() *CacheStats {
	return createCacheStats(0, 0, 0)
}

func createCacheStats(hitCount uint64, missCount uint64, evictionCount uint64) *CacheStats {
	cs := &CacheStats{
		hitCount:      new(uint64),
		missCount:     new(uint64),
		evictionCount: new(uint64),
	}
	*cs.hitCount = hitCount
	*cs.missCount = missCount
	*cs.evictionCount = evictionCount
	return cs
}

func (s *CacheStats) HitCount() uint64 {
	return atomic.LoadUint64(s.hitCount)
}

func (s *CacheStats) MissCount() uint64 {
	return atomic.LoadUint64(s.missCount)
}

func (s *CacheStats) RequestCount() uint64 {
	return s.HitCount() + s.MissCount()
}

func (s *CacheStats) EvictionCount() uint64 {
	return atomic.LoadUint64(s.evictionCount)
}

func (s *CacheStats) HitRate() float64 {
	requestCount := s.RequestCount()
	if requestCount == 0 {
		return 1.0
	}
	return float64(s.HitCount()) / float64(requestCount)
}

func (s *CacheStats) MissRate() float64 {
	requestCount := s.RequestCount()
	if requestCount == 0 {
		return 0.0
	}
	return float64(s.MissCount()) / float64(requestCount)
}

func (s *CacheStats) RecordHits() {
	atomic.AddUint64(s.hitCount, 1)
}

func (s *CacheStats) RecordMisses() {
	atomic.AddUint64(s.missCount, 1)
}

func (s *CacheStats) RecordEviction() {
	atomic.AddUint64(s.evictionCount, 1)
}

func (s *CacheStats) Snapshot() *CacheStats {
	return createCacheStats(s.HitCount(), s.MissCount(), s.EvictionCount())
}
