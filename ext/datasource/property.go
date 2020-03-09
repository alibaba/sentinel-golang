package datasource

import (
	"reflect"
)

// PropertyConvert func is to convert source message bytes to the specific property.
type PropertyConvert func(src []byte) interface{}

// PropertyUpdate func is to update the specific properties to downstream.
type PropertyUpdate func(data interface{}) error

// abstract interface to
type PropertyHandler interface {
	// check whether the current src is consistent with last update property
	isPropertyConsistent(src interface{}) bool
	// handle the current property
	Handle(src []byte) error
}

// DefaultPropertyHandler encapsulate the Converter and updater of property.
// One DefaultPropertyHandler instance is to handle one property type.
// DefaultPropertyHandler should check whether current property is consistent with last update property
// converter convert the message to the specific property
// updater update the specific property to downstream.
type DefaultPropertyHandler struct {
	lastUpdateProperty interface{}

	convert PropertyConvert
	update  PropertyUpdate
}

func (h *DefaultPropertyHandler) isPropertyConsistent(src interface{}) bool {
	isConsistent := reflect.DeepEqual(src, h.lastUpdateProperty)
	if isConsistent {
		return true
	} else {
		h.lastUpdateProperty = src
		return false
	}
}

func (h *DefaultPropertyHandler) Handle(src []byte) error {
	// convert to target property
	realProperty := h.convert(src)
	isConsistent := h.isPropertyConsistent(realProperty)
	if isConsistent {
		return nil
	}
	return h.update(realProperty)
}

func NewSinglePropertyHandler(convert PropertyConvert, update PropertyUpdate) *DefaultPropertyHandler {
	return &DefaultPropertyHandler{
		convert: convert,
		update:  update,
	}
}
