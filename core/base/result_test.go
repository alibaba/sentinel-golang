package base

import (
	"testing"
)

func BenchmarkWithPool(b *testing.B) {
	var result *TokenResult
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
			result = NewTokenResultPass()
			RefurbishTokenResult(result)
		}
	}
}

func BenchmarkWithNoPool(b *testing.B) {
	var result *TokenResult
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := 0; j < 10000; j++ {
			result = NewTokenResultEmpty()
			result.status = ResultStatusPass
			result.blockErr = nil
			result.waitMs = 0
		}
	}
}
