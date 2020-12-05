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

package base

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var flag = 0

func exitHandlerMock(entry *SentinelEntry, ctx *EntryContext) error {
	flag += 1
	return nil
}

func TestSentinelEntry_WhenExit(t *testing.T) {
	flag = 0
	sc := NewSlotChain()
	ctx := sc.GetPooledContext()
	entry := NewSentinelEntry(ctx, nil, sc)
	entry.WhenExit(exitHandlerMock)
	entry.Exit()
	assert.True(t, flag == 1)

	entry.Exit()
	assert.True(t, flag == 1)
}
