package cache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountMinSketch(t *testing.T) {
	t.Run("Test_CountMinSketch", func(t *testing.T) {
		max := 15
		cm4 := newCountMinSketch(max)
		for i := 0; i < max; i++ {
			for j := i; j > 0; j-- {
				cm4.add(uint64(i))
			}
			assert.True(t, uint64(i) == uint64(cm4.estimate(uint64(i))))
		}

		cm4.reset()
		for i := 0; i < max; i++ {
			assert.True(t, uint64(i)/2 == uint64(cm4.estimate(uint64(i))))
		}

		cm4.clear()
		for i := 0; i < max; i++ {
			assert.True(t, 0 == uint64(cm4.estimate(uint64(i))))
		}
	})
}
