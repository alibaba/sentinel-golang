package util

import "math"

const precision = 0.00000001

func Float64Equals(x, y float64) bool {
	return math.Abs(x-y) < precision
}
