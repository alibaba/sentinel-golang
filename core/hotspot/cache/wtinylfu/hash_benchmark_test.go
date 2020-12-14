package wtinylfu

import (
	"testing"
)

func Benchmark_Hash_Num(b *testing.B) {
	num := 100020000
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum(num)
	}
}
func Benchmark_Hash_String(b *testing.B) {
	str := "test"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum(str)
	}
}

func Benchmark_Hash_Pointer(b *testing.B) {
	num := 100020000
	pointer := &num
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum(pointer)
	}
}
func Benchmark_Hash_WithSprintf(b *testing.B) {
	type test struct {
		test1 uint32
		test2 string
	}
	s := test{
		1,
		"test2222",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sum(s)
	}
}
