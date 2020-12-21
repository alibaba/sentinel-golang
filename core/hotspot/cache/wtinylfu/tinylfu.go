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
	"container/list"
	"errors"

	"github.com/alibaba/sentinel-golang/core/hotspot/cache/stats"
)

const (
	samplesFactor            = 8
	doorkeeperFactor         = 8
	countersFactor           = 1
	falsePositiveProbability = 0.1
	lruRatio                 = 0.01
)

// TinyLfu is an implementation of the TinyLfu algorithm: https://arxiv.org/pdf/1512.00727.pdf
//           Window Cache Victim .---------. Main Cache Victim
//          .------------------->| TinyLFU |<-----------------.
//          |                    `---------'                  |
// .-------------------.              |    .------------------.
// | Window Cache (1%) |              |    | Main Cache (99%) |
// |      (LRU)        |              |    |      (SLRU)      |
// `-------------------'              |    `------------------'
//          ^                         |               ^
//          |                         `---------------'
//       new item                        Winner
type TinyLfu struct {
	countMinSketch *countMinSketch
	doorkeeper     *doorkeeper
	additions      int
	samples        int
	lru            *lru
	slru           *slru
	items          map[interface{}]*list.Element
	stats          *stats.CacheStats
}

func NewTinyLfu(cap int, isRecordingStats bool) (*TinyLfu, error) {
	if cap <= 0 {
		return nil, errors.New("Must provide a positive size")
	}
	if cap < 100 {
		cap = 100
	}
	lruCap := int(float64(cap) * lruRatio)
	slruSize := cap - lruCap
	var statsCache *stats.CacheStats
	if isRecordingStats {
		statsCache = stats.NewCacheStats()
	}
	items := make(map[interface{}]*list.Element)
	return &TinyLfu{
		countMinSketch: newCountMinSketch(countersFactor * cap),
		additions:      0,
		samples:        samplesFactor * cap,
		doorkeeper:     newDoorkeeper(doorkeeperFactor*cap, falsePositiveProbability),
		items:          items,
		lru:            newLRU(lruCap, items),
		slru:           newSLRU(slruSize, items),
		stats:          statsCache,
	}, nil
}

// Get looks up a key's value from the cache.
func (t *TinyLfu) Get(key interface{}) (interface{}, bool) {
	return t.get(key, false)
}

func (t *TinyLfu) get(key interface{}, isInternal bool) (interface{}, bool) {
	t.additions++
	if t.additions == t.samples {
		t.countMinSketch.reset()
		t.doorkeeper.reset()
		t.additions = 0
	}

	val, ok := t.items[key]
	if !ok {
		keyHash := sum(key)
		if t.doorkeeper.put(keyHash) {
			t.countMinSketch.add(keyHash)
		}
		if !isInternal && t.stats != nil {
			t.stats.RecordMisses()
		}
		return nil, false
	}
	item := val.Value.(*slruItem)
	if t.doorkeeper.put(item.keyHash) {
		t.countMinSketch.add(item.keyHash)
	}

	v := item.value
	if item.listId == admissionWindow {
		t.lru.get(val)
	} else {
		t.slru.get(val)
	}
	if !isInternal && t.stats != nil {
		t.stats.RecordHits()
	}
	return v, true
}

// Contains checks if a key is in the cache without updating
func (t *TinyLfu) Contains(key interface{}) (ok bool) {
	_, ok = t.items[key]
	return ok
}

func (t *TinyLfu) Add(key interface{}, val interface{}) {
	t.AddIfAbsent(key, val)
}

// AddIfAbsent adds item only if key is not existed.
func (t *TinyLfu) AddIfAbsent(key interface{}, val interface{}) (priorValue interface{}) {

	// Check for existing item
	if v, ok := t.get(key, true); ok {
		return v
	}

	newItem := slruItem{admissionWindow, key, val, sum(key)}
	candidate, evicted := t.lru.add(newItem)
	if !evicted {
		return nil
	}

	// Estimate count of what will be evicted from slru
	victim := t.slru.victim()
	if victim == nil {
		t.slru.add(candidate)
		return nil
	}

	victimCount := t.estimate(victim.keyHash)
	candidateCount := t.estimate(candidate.keyHash)
	if candidateCount > victimCount {
		t.slru.add(candidate)
	}
	if t.stats != nil {
		t.stats.RecordEviction()
	}
	return nil
}

// estimate estimates frequency of the given hash value.
func (t *TinyLfu) estimate(h uint64) uint8 {
	freq := t.countMinSketch.estimate(h)
	if t.doorkeeper.contains(h) {
		freq++
	}
	return freq
}

func (t *TinyLfu) Remove(key interface{}) (isFound bool) {
	// Check for existing item
	val, ok := t.items[key]
	if !ok {
		return false
	}

	item := val.Value.(*slruItem)
	if item.listId == admissionWindow {
		t.lru.Remove(key)
		return true
	} else {
		t.slru.Remove(key)
		return true
	}
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (t *TinyLfu) Keys() []interface{} {
	i := 0
	keys := make([]interface{}, len(t.items))
	for k := range t.items {
		keys[i] = k
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (t *TinyLfu) Len() int {
	return len(t.items)
}

// Purge is used to completely clear the cache.
func (t *TinyLfu) Purge() {
	for k := range t.items {
		delete(t.items, k)
	}
	t.slru.clear()
	t.additions = 0
	t.lru.clear()
	t.doorkeeper.reset()
	t.countMinSketch.clear()
}

// Stats copies cache stats.
func (t *TinyLfu) Stats() (*stats.CacheStats, error) {
	if t.stats == nil {
		return nil, errors.New("RecordingStats Must be enabled")
	}
	return t.stats.Snapshot(), nil
}
