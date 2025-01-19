package statlogger

import (
	"testing"
)

func Benchmark_StatLogger_Stat(b *testing.B) {
	testLogger := NewStatLogger("test-sentinel.log", 3, 1000, 6000, 200000000)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testLogger.Stat(1, "test", "test2", "test3")
	}
}
