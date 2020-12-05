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
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/stretchr/testify/assert"
)

const (
	systemRule = `[
    {
        "metricType": 0,
        "triggerCount": 0.5,
        "strategy": 1
    }
]`
	systemRuleErr = "[test]"
)

func TestBase_AddPropertyHandler(t *testing.T) {
	t.Run("AddPropertyHandler_nil", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		b.AddPropertyHandler(nil)
		assert.True(t, len(b.handlers) == 0, "Fail to execute the case TestBase_AddPropertyHandler.")
	})

	t.Run("AddPropertyHandler_Normal", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		h := &DefaultPropertyHandler{}
		b.AddPropertyHandler(h)
		assert.True(t, len(b.handlers) == 1 && reflect.DeepEqual(b.handlers[0], h), "Fail to execute the case TestBase_AddPropertyHandler.")
	})
}

func TestBase_RemovePropertyHandler(t *testing.T) {
	t.Run("TestBase_RemovePropertyHandler_nil", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		h1 := &DefaultPropertyHandler{}
		b.handlers = append(b.handlers, h1)
		b.RemovePropertyHandler(nil)
		assert.True(t, len(b.handlers) == 1, "The case TestBase_RemovePropertyHandler execute failed.")
	})

	t.Run("TestBase_RemovePropertyHandler", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		h1 := &DefaultPropertyHandler{}
		b.handlers = append(b.handlers, h1)
		b.RemovePropertyHandler(h1)
		assert.True(t, len(b.handlers) == 0, "The case TestBase_RemovePropertyHandler execute failed.")
	})
}

func TestBase_indexOfHandler(t *testing.T) {
	t.Run("TestBase_indexOfHandler", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		h1 := &DefaultPropertyHandler{}
		b.handlers = append(b.handlers, h1)
		h2 := &DefaultPropertyHandler{}
		b.handlers = append(b.handlers, h2)
		h3 := &DefaultPropertyHandler{}
		b.handlers = append(b.handlers, h3)

		assert.True(t, b.indexOfHandler(h2) == 1, "Fail to execute the case TestBase_indexOfHandler.")
	})
}

func TestBase_Handle(t *testing.T) {
	t.Run("TestBase_handle", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		h := NewSystemRulesHandler(SystemRuleJsonArrayParser)
		b.handlers = append(b.handlers, h)
		err := b.Handle([]byte(systemRule))
		assert.Nil(t, err)
		assert.True(t, len(system.GetRules()) == 1)
	})

	t.Run("TestBase_multipleHandle", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		systemHandler := NewSystemRulesHandler(SystemRuleJsonArrayParser)
		flowHandler := NewFlowRulesHandler(FlowRuleJsonArrayParser)
		b.handlers = append(b.handlers, systemHandler)
		b.handlers = append(b.handlers, flowHandler)
		err := b.Handle([]byte(systemRule))
		assert.Nil(t, err)
		assert.True(t, len(system.GetRules()) == 1)
		assert.True(t, len(flow.GetRules()) == 0)
	})

	t.Run("TestBase_handleErr", func(t *testing.T) {
		b := &Base{
			handlers: make([]PropertyHandler, 0),
		}
		h := NewSystemRulesHandler(SystemRuleJsonArrayParser)
		b.handlers = append(b.handlers, h)
		err := b.Handle([]byte(systemRuleErr))
		assert.NotNil(t, err)
	})
}
