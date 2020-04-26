package util

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicBool_CompareAndSet(t *testing.T) {
	b := &AtomicBool{}
	b.Set(true)
	ok := b.CompareAndSet(true, false)
	assert.True(t, ok, "CompareAndSet execute failed.")
	b.Set(false)
	ok = b.CompareAndSet(true, false)
	assert.True(t, !ok, "CompareAndSet execute failed.")
}

func TestAtomicBool_GetAndSet(t *testing.T) {
	b := &AtomicBool{}
	assert.True(t, b.Get() == false, "default value is not false.")
	b.Set(true)
	assert.True(t, b.Get() == true, "the value is false, expect true.")
}

func TestIncrementAndGetInt64(t *testing.T) {
	n := int64(0)
	wg := &sync.WaitGroup{}
	wg.Add(100)
	for i := 0; i < 100; i++ {
		go func(g *sync.WaitGroup) {
			IncrementAndGetInt64(&n)
			wg.Done()
		}(wg)
	}
	wg.Wait()
	assert.True(t, n == 100, fmt.Sprintf("current n is %d, expect 100.", n))
}
