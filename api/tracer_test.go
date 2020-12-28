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

package api

import (
	"errors"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/stretchr/testify/assert"
)

var (
	testRes = base.NewResourceWrapper("a", base.ResTypeCommon, base.Inbound)
)

func TestTraceErrorToEntry(t *testing.T) {
	util.SetClock(util.NewMockClock())

	type args struct {
		entry *base.SentinelEntry
		err   error
	}
	te := errors.New("biz error")
	tests := []struct {
		name string
		args args
		want error
	}{
		{
			name: "TestTraceErrorToEntry",
			args: args{
				entry: nil,
				err:   nil,
			},
			want: te,
		},
	}

	ctx := &base.EntryContext{
		Resource: testRes,
		Input:    nil,
	}
	tests[0].args.entry = base.NewSentinelEntry(ctx, testRes, nil)
	tests[0].args.err = te

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			TraceError(tt.args.entry, tt.args.err)
			util.Sleep(time.Millisecond * 10)
			assert.Equal(t, tests[0].args.entry.Context().Err(), tt.want)
		})
	}
}
