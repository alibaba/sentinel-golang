package wtinylfu

import (
	"sync"

	"github.com/alibaba/sentinel-golang/core/hotspot/cache"
)

// TinyLfuCacheMap use tinyLfu strategy to cache the most frequently accessed hotspot parameter
type TinyLfuCacheMap struct {
	// Not thread safe
	tinyLfu *TinyLfu
	lock    *sync.RWMutex
}

func (c *TinyLfuCacheMap) Add(key interface{}, value *int64) {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tinyLfu.Add(key, value)
	return
}

func (c *TinyLfuCacheMap) AddIfAbsent(key interface{}, value *int64) (priorValue *int64) {
	c.lock.Lock()
	defer c.lock.Unlock()
	val := c.tinyLfu.AddIfAbsent(key, value)
	if val == nil {
		return nil
	}
	priorValue = val.(*int64)
	return
}

func (c *TinyLfuCacheMap) Get(key interface{}) (value *int64, isFound bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	val, found := c.tinyLfu.Get(key)
	if found {
		return val.(*int64), true
	}
	return nil, false
}

func (c *TinyLfuCacheMap) Remove(key interface{}) (isFound bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	return c.tinyLfu.Remove(key)
}

func (c *TinyLfuCacheMap) Contains(key interface{}) (ok bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.tinyLfu.Contains(key)
}

func (c *TinyLfuCacheMap) Keys() []interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.tinyLfu.Keys()
}

func (c *TinyLfuCacheMap) Len() int {
	c.lock.RLock()
	defer c.lock.RUnlock()

	return c.tinyLfu.Len()
}

func (c *TinyLfuCacheMap) Purge() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.tinyLfu.Purge()
}

func NewTinyLfuCacheMap(size int) cache.ConcurrentCounterCache {
	tinyLfu, err := NewTinyLfu(size)
	if err != nil {
		return nil
	}
	return &TinyLfuCacheMap{
		tinyLfu: tinyLfu,
		lock:    new(sync.RWMutex),
	}
}
