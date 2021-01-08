// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
func Benchmark_Hash_OtherWithSprintf(b *testing.B) {
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
