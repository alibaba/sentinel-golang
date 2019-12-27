package datasource

import (
	"reflect"
	"sync"

	"github.com/sentinel-group/sentinel-golang/logging"
)

const (
	FlagInitialLoaded int32 = 1
	FlagUpdate int32 = 2
	FlagDelete int32 = 3
)

var logger = logging.GetDefaultLogger()

type listenerSet map[PropertyListener]struct{}

type PropertyListener interface {
	OnConfigUpdate(value interface{}, flag int32) error
}

type PropertyPublisher interface {
	UpdateValue(value interface{}, flag int32) (bool, error)

	AddListener(listener PropertyListener) bool
	RemoveListener(listener PropertyListener)
}

// SentinelProperty represents the publisher.
type SentinelProperty struct {
	listeners listenerSet
	value     interface{}

	listenerMux *sync.RWMutex
	valueMux    *sync.RWMutex
}

func NewSentinelProperty() *SentinelProperty {
	return &SentinelProperty{
		listeners:   make(listenerSet),
		listenerMux: new(sync.RWMutex),
		valueMux:    new(sync.RWMutex),
	}
}

func (s *SentinelProperty) UpdateValue(v interface{}, flag int32) (bool, error) {
	s.valueMux.Lock()
	defer s.valueMux.Unlock()

	if reflect.DeepEqual(v, s.value) {
		return false, nil
	}
	s.value = v

	s.listenerMux.RLock()
	defer s.listenerMux.RUnlock()

	var err error
	for listener := range s.listeners {
		err = listener.OnConfigUpdate(v, flag)
		if err != nil {
			logger.Warnf("Error when updating data value: %+v", err)
		}
	}
	return true, err
}

func (s *SentinelProperty) AddListener(listener PropertyListener) bool {
	if listener == nil {
		return false
	}
	s.listenerMux.Lock()
	defer s.listenerMux.Unlock()

	if _, exists := s.listeners[listener]; exists {
		return false
	}
	s.listeners[listener] = struct{}{}
	return true
}

func (s *SentinelProperty) RemoveListener(listener PropertyListener) {
	if listener == nil {
		return
	}
	s.listenerMux.Lock()
	defer s.listenerMux.Unlock()

	delete(s.listeners, listener)
}
