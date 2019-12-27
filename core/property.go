package core

import (
	"go.uber.org/multierr"
)

// Listener
type Listener interface {
	ConfigUpdate([]byte) error
	ConfigLoad([]byte) error
}

type DynamicSentinelProperty struct {
	listeners []Listener
}

func NewDynamicSentinelProperty() *DynamicSentinelProperty {
	return &DynamicSentinelProperty{listeners:make([]Listener, 0)}
}

func (dsp DynamicSentinelProperty) UpdateValue(data []byte) (err error) {
	for _, listener := range dsp.listeners {
		err = multierr.Append(err, listener.ConfigUpdate(data))
	}
	return
}

func (dsp DynamicSentinelProperty) SetValue(data []byte) (err error) {
	for _, listener := range dsp.listeners {
		err = multierr.Append(err, listener.ConfigLoad(data))
	}
	return
}

func (dsp *DynamicSentinelProperty) AddListener(listener Listener) {
	dsp.listeners = append(dsp.listeners, listener)
}