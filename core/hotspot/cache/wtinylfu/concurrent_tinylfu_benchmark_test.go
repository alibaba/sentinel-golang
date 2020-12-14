package wtinylfu

import (
	"strconv"
	"testing"
)

const Size = 50000

func Benchmark_TINYLFU_AddIfAbsent(b *testing.B) {
	c := NewTinyLfuCacheMap(Size)
	for a := 1; a <= Size; a++ {
		val := new(int64)
		*val = int64(a)
		c.Add(strconv.Itoa(a), val)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 1000; j <= 1001; j++ {
			newVal := new(int64)
			*newVal = int64(j)
			prior := c.AddIfAbsent(strconv.Itoa(j), newVal)
			if *prior != int64(j) {
				b.Fatal("error!")
			}
		}
	}
}

func Benchmark_TINYLFU_Add(b *testing.B) {
	c := NewTinyLfuCacheMap(Size)
	for a := 1; a <= Size; a++ {
		val := new(int64)
		*val = int64(a)
		c.Add(strconv.Itoa(a), val)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := Size - 100; j <= Size-99; j++ {
			newVal := new(int64)
			*newVal = int64(j)
			c.Add(strconv.Itoa(j), newVal)
		}
	}
}

func Benchmark_TINYLFU_Get(b *testing.B) {
	c := NewTinyLfuCacheMap(Size)
	for a := 1; a <= Size; a++ {
		val := new(int64)
		*val = int64(a)
		c.Add(strconv.Itoa(a), val)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := Size - 100; j <= Size-99; j++ {
			val, found := c.Get(strconv.Itoa(j))
			if *val != int64(j) || !found {
				b.Fatal("error")
			}
		}
	}
}
