package wtinylfu

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDoorkeeper(t *testing.T) {
	t.Run("Test_Doorkeeper", func(t *testing.T) {
		max := 1500
		filter := newDoorkeeper(1500, 0.001)
		for i := 0; i < max; i++ {
			filter.put(uint64(i))
			assert.True(t, true == filter.contains(uint64(i)))
		}
		filter.reset()
		for i := 0; i < max; i++ {
			assert.True(t, false == filter.contains(uint64(i)))
		}
	})
}
