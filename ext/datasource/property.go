package datasource

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/pkg/errors"
	"reflect"
)

var logger = logging.GetDefaultLogger()

// PropertyConverter func is to converter source message bytes to the specific property.
type PropertyConverter func(src []byte) interface{}

// PropertyUpdater func is to update the specific properties to downstream.
type PropertyUpdater func(data interface{}) error

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
		if err := recover(); err != nil && logger != nil {
			logger.Panicf("Unexpected panic: %+v", errors.Errorf("%+v", err))
		}
	}()
	// converter to target property
	realProperty := h.converter(src)
	isConsistent := h.isPropertyConsistent(realProperty)
	if isConsistent {
		return nil
	}
	return h.updater(realProperty)
}

func NewSinglePropertyHandler(converter PropertyConverter, updater PropertyUpdater) *DefaultPropertyHandler {
	return &DefaultPropertyHandler{
		converter: converter,
		updater:   updater,
	}
}
