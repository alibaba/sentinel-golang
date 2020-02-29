package base

import (
	"io"
	"sync"

	"go.uber.org/multierr"
)

var propertyConsumers []func(decoder PropertyDecoder)error
var propertyResets []func() error
var registerPropertyConsumerMutex sync.Mutex

// RegisterPropertyConsumer register callback for property change event.
// When property updated, consumer will be called to build an decoder,
// which decode property to *Rules(FlowRules/SystemRules etc.)
// When property deleted, reset will be called to delete *Rules.
func RegisterPropertyConsumer(consumer func(decoder PropertyDecoder) error, reset func() error) {
	registerPropertyConsumerMutex.Lock()
	defer registerPropertyConsumerMutex.Unlock()

	propertyConsumers = append(propertyConsumers, consumer)
	propertyResets = append(propertyResets, reset)
}

// UpdateProperty update property. Called by `DataSource` when property updated.
func UpdateProperty(builder func() PropertyDecoder) (err error) {
	for _, consumer := range propertyConsumers {
		err = multierr.Append(err, consumer(builder()))
	}
	return err
}

// DeleteProperty delete property. Called by `DataSource` when property deleted.
func DeleteProperty() (err error) {
	for _, reset := range propertyResets {
		err = multierr.Append(err, reset())
	}
	return err
}

// DataSource provides ReadConfig method to read config,
// and Close method to release resource.
type DataSource interface {
	ReadConfig() error
	io.Closer
}

// PropertyDecoder decode property to rule
type PropertyDecoder interface {
	Decode(interface{}) error
}

