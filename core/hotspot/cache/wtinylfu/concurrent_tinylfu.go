package wtinylfu

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/hotspot/cache"
)

// TinyLfuCacheMap use tinyLfu strategy to cache the most frequently accessed hotspot parameter
type TinyLfuCacheMap struct {
	// Not thread safe
	tinyLfu *TinyLfu
	sync.RWMutex
}

func (c *TinyLfuCacheMap) Add(key interface{}, value *int64) {
	c.Lock()
	defer c.Unlock()

	c.tinyLfu.Add(key, value)
	return
}

func (c *TinyLfuCacheMap) AddIfAbsent(key interface{}, value *int64) (priorValue *int64) {
	c.Lock()
	defer c.Unlock()
	val := c.tinyLfu.AddIfAbsent(key, value)
	if val == nil {
		return nil
	}
	priorValue = val.(*int64)
	return
}

func (c *TinyLfuCacheMap) Get(key interface{}) (value *int64, isFound bool) {
	c.Lock()
	defer c.Unlock()

	val, found := c.tinyLfu.Get(key)
	if found {
		return val.(*int64), true
	}
	return nil, false
}

func (c *TinyLfuCacheMap) Remove(key interface{}) (isFound bool) {
	c.Lock()
	defer c.Unlock()

	return c.tinyLfu.Remove(key)
}

func (c *TinyLfuCacheMap) Contains(key interface{}) (ok bool) {
	c.RLock()
	defer c.RUnlock()

	return c.tinyLfu.Contains(key)
}

func (c *TinyLfuCacheMap) Keys() []interface{} {
	c.RLock()
	defer c.RUnlock()

	return c.tinyLfu.Keys()
}

func (c *TinyLfuCacheMap) Len() int {
	c.RLock()
	defer c.RUnlock()

	return c.tinyLfu.Len()
}

func (c *TinyLfuCacheMap) Purge() {
	c.Lock()
	defer c.Unlock()

	c.tinyLfu.Purge()
}

func NewTinyLfuCacheMap(size int) cache.ConcurrentCounterCache {
	tinyLfu, err := NewTinyLfu(size)
	if err != nil {
		return nil
	}
	return &TinyLfuCacheMap{
		tinyLfu: tinyLfu,
	}
}
