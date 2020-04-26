package datasource

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
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
