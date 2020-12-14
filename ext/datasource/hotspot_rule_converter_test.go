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

package datasource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_parseSpecificItems(t *testing.T) {
	t.Run("Test_parseSpecificItems", func(t *testing.T) {
		source := make([]SpecificValue, 6)
		s1 := SpecificValue{
			ValKind:   KindInt,
			ValStr:    "10010",
			Threshold: 100,
		}
		s2 := SpecificValue{
			ValKind:   KindInt,
			ValStr:    "10010aaa",
			Threshold: 100,
		}
		s3 := SpecificValue{
			ValKind:   KindString,
			ValStr:    "test-string",
			Threshold: 100,
		}
		s4 := SpecificValue{
			ValKind:   KindBool,
			ValStr:    "true",
			Threshold: 100,
		}
		s5 := SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.234",
			Threshold: 100,
		}
		s6 := SpecificValue{
			ValKind:   KindFloat64,
			ValStr:    "1.2345678",
			Threshold: 100,
		}
		source[0] = s1
		source[1] = s2
		source[2] = s3
		source[3] = s4
		source[4] = s5
		source[5] = s6

		got := parseSpecificItems(source)
		assert.True(t, len(got) == 5)
		assert.True(t, got[10010] == 100)
		assert.True(t, got[true] == 100)
		assert.True(t, got[1.234] == 100)
		assert.True(t, got[1.23400] == 100)
		assert.True(t, got["test-string"] == 100)
		assert.True(t, got[1.23457] == 100)
	})
}
