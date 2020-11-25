package cache

import "container/list"

const (
	admissionWindow uint8 = iota
	probationSegment
	protectedSegment
)

const protectedRatio = 0.8

type slruItem struct {
	listId  uint8
	key     interface{}
	value   interface{}
	keyHash uint64
}

// slru is a segmented LRU.
type slru struct {
	data                       map[interface{}]*list.Element
	probationCap, protectedCap int
	probationLs, protectedLs   *list.List
}

func newSLRU(cap int, data map[interface{}]*list.Element) *slru {
	protectedCap := int(float64(cap) * protectedRatio)
	probationCap := cap - protectedCap
	return &slru{
		data:         data,
		probationCap: probationCap,
		probationLs:  list.New(),
		protectedCap: protectedCap,
		protectedLs:  list.New(),
	}
}

// Get looks up a key's value from the cache.
func (slru *slru) get(v *list.Element) {
	item := v.Value.(*slruItem)
	if item.listId == protectedSegment {
		slru.protectedLs.MoveToFront(v)
		return
	}
	if slru.protectedLs.Len() < slru.protectedCap {
		slru.probationLs.Remove(v)
		item.listId = protectedSegment
		slru.data[item.key] = slru.protectedLs.PushFront(item)
		return
	}
	back := slru.protectedLs.Back()
	backItem := back.Value.(*slruItem)

	// swap the two item
	*backItem, *item = *item, *backItem
	backItem.listId = protectedSegment
	item.listId = probationSegment
	slru.data[item.key] = v
	slru.data[backItem.key] = back

	// move the elements to the front of their lists
	slru.probationLs.MoveToFront(v)
	slru.protectedLs.MoveToFront(back)
}

// add set a value in the cache
func (slru *slru) add(newItem slruItem) {
	newItem.listId = probationSegment
	if slru.probationLs.Len() < slru.probationCap || slru.Len() < slru.probationCap+slru.protectedCap {
		slru.data[newItem.key] = slru.probationLs.PushFront(&newItem)
		return
	}
	back := slru.probationLs.Back()
	item := back.Value.(*slruItem)
	delete(slru.data, item.key)
	*item = newItem
	slru.data[item.key] = back
	slru.probationLs.MoveToFront(back)
}

func (slru *slru) victim() *slruItem {
	if slru.Len() < slru.probationCap+slru.protectedCap {
		return nil
	}
	v := slru.probationLs.Back()
	return v.Value.(*slruItem)
}

// Len returns the total number of items in the cache
func (slru *slru) Len() int {
	return slru.probationLs.Len() + slru.protectedLs.Len()
}

// Remove removes an item from the cache, returning the item and a boolean indicating if it was found
func (slru *slru) Remove(key interface{}) (interface{}, bool) {
	v, ok := slru.data[key]
	if !ok {
		return nil, false
	}
	item := v.Value.(*slruItem)
	if item.listId == protectedSegment {
		slru.protectedLs.Remove(v)
	} else {
		slru.probationLs.Remove(v)
	}
	delete(slru.data, key)
	return item.value, true
}

func (slru *slru) clear() {
	slru.probationLs.Init()
	slru.protectedLs.Init()
}

// lru is an LRU cache.
type lru struct {
	data      map[interface{}]*list.Element
	cap       int
	evictList *list.List
}

func newLRU(cap int, data map[interface{}]*list.Element) *lru {
	return &lru{
		data:      data,
		cap:       cap,
		evictList: list.New(),
	}
}

// Get returns a value from the cache
func (lru *lru) get(v *list.Element) {
	lru.evictList.MoveToFront(v)
}

// Set a value in the cache
func (lru *lru) add(newItem slruItem) (oldItem slruItem, evicted bool) {
	if lru.evictList.Len() < lru.cap {
		lru.data[newItem.key] = lru.evictList.PushFront(&newItem)
		return slruItem{}, false
	}

	// reuse the item
	e := lru.evictList.Back()
	item := e.Value.(*slruItem)

	delete(lru.data, item.key)

	oldItem = *item
	*item = newItem

	lru.data[item.key] = e
	lru.evictList.MoveToFront(e)

	return oldItem, true
}

// Len returns the number of items in the cache.
func (lru *lru) Len() int {
	return lru.evictList.Len()
}

// Remove removes the provided key from the cache
func (lru *lru) Remove(key interface{}) (interface{}, bool) {
	v, ok := lru.data[key]
	if !ok {
		return nil, false
	}
	item := v.Value.(*slruItem)
	lru.evictList.Remove(v)
	delete(lru.data, key)
	return item.value, true
}

func (lru *lru) clear() {
	lru.evictList.Init()
}
