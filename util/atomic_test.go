package util

import (
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
