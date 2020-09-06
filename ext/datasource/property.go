package datasource

import (
	"reflect"

	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
)

// PropertyConverter func is to convert source message bytes to the specific property.
// the first  return value: is the real property;
// the second return value: return nil if succeed to convert src, if not return the detailed error when convert src.
// if src is nil or len(src)==0, the return value is (nil,nil)
type PropertyConverter func(src []byte) (interface{}, error)

// PropertyUpdater func is to update the specific properties to downstream.
// return nil if succeed to update, if not, return the error.
type PropertyUpdater func(data interface{}) error

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

	converter PropertyConverter
	updater   PropertyUpdater
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
	defer func() {
		if err := recover(); err != nil {
			logging.Panicf("Unexpected panic: %+v", errors.Errorf("%+v", err))
		}
	}()
	// convert to target property
	realProperty, err := h.converter(src)
	if err != nil {
		return err
	}
	isConsistent := h.isPropertyConsistent(realProperty)
	if isConsistent {
		return nil
	}
	return h.updater(realProperty)
}

func NewDefaultPropertyHandler(converter PropertyConverter, updater PropertyUpdater) *DefaultPropertyHandler {
	return &DefaultPropertyHandler{
		converter: converter,
		updater:   updater,
	}
}
