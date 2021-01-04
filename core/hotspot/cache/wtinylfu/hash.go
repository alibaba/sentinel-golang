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
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"math"
	"reflect"
)

var (
	fnv64   = fnv.New64()
	byteSum = make([]byte, 0, 8)
)

func sum(k interface{}) uint64 {
	switch h := k.(type) {
	case int:
		return hashU64(uint64(h))
	case int8:
		return hashU64(uint64(h))
	case int16:
		return hashU64(uint64(h))
	case int32:
		return hashU64(uint64(h))
	case int64:
		return hashU64(uint64(h))
	case uint:
		return hashU64(uint64(h))
	case uint8:
		return hashU64(uint64(h))
	case uint16:
		return hashU64(uint64(h))
	case uint32:
		return hashU64(uint64(h))
	case uint64:
		return hashU64(h)
	case uintptr:
		return hashU64(uint64(h))
	case float32:
		return hashU64(uint64(math.Float32bits(h)))
	case float64:
		return hashU64(math.Float64bits(h))
	case bool:
		if h {
			return 1
		}
		return 0
	case string:
		return hashString(h)
	}
	if h, ok := hashPointer(k); ok {
		return h
	}
	if h, ok := hashOtherWithSprintf(k); ok {
		return h
	}
	return 0
}

func hashU64(data uint64) uint64 {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, data)
	return hashByteArray(b)
}

func hashString(data string) uint64 {
	return hashByteArray([]byte(data))
}

func hashOtherWithSprintf(data interface{}) (uint64, bool) {
	v := fmt.Sprintf("%v", data)
	return hashString(v), true
}

func hashByteArray(bytes []byte) uint64 {
	_, err := fnv64.Write(bytes)
	if err != nil {
		return 0
	}
	hash := binary.LittleEndian.Uint64(fnv64.Sum(byteSum))
	fnv64.Reset()
	return hash
}

func hashPointer(k interface{}) (uint64, bool) {
	v := reflect.ValueOf(k)
	switch v.Kind() {
	case reflect.Ptr, reflect.UnsafePointer, reflect.Func, reflect.Slice, reflect.Map, reflect.Chan:
		return hashU64(uint64(v.Pointer())), true
	default:
		return 0, false
	}
}
