package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFloat64Equals(t *testing.T) {
	assert.True(t, Float64Equals(0.1, 0.099999999))
	assert.False(t, Float64Equals(0.1, 0.09999999))
}
