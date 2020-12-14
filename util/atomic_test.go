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

package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAtomicBool_CompareAndSet(t *testing.T) {
	b := &AtomicBool{}
	b.Set(true)
	ok := b.CompareAndSet(true, false)
	assert.True(t, ok, "CompareAndSet execute failed.")
	b.Set(false)
	ok = b.CompareAndSet(true, false)
	assert.True(t, !ok, "CompareAndSet execute failed.")
}

func TestAtomicBool_GetAndSet(t *testing.T) {
	b := &AtomicBool{}
	assert.True(t, b.Get() == false, "default value is not false.")
	b.Set(true)
	assert.True(t, b.Get() == true, "the value is false, expect true.")
}
