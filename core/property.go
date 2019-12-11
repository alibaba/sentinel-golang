package core

import "golang.org/x/sync/errgroup"

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

func (dsp DynamicSentinelProperty) UpdateValue(data []byte) error {
	var eg errgroup.Group
	for _, listener := range dsp.listeners {
		eg.Go(func() error {
			return listener.ConfigUpdate(data)
		})
	}
	return eg.Wait()
}

func (dsp DynamicSentinelProperty) SetValue(data []byte) error {
	var eg errgroup.Group
	for _, listener := range dsp.listeners {
		eg.Go(func() error {
			return listener.ConfigLoad(data)
		})
	}
	return eg.Wait()
}

func (dsp *DynamicSentinelProperty) AddListener(listener Listener) {
	dsp.listeners = append(dsp.listeners, listener)
}