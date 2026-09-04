// Copyright 1999-2026 Alibaba Group Holding Ltd.
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

package metric

import (
	"net/http"
	"sync"
)

type bindableMetric interface {
	bind(Exporter) error
}

// deferredExporter records metrics declared by package initializers and binds
// them once the application configuration is available.
type deferredExporter struct {
	mu       sync.RWMutex
	delegate Exporter
	metrics  []bindableMetric
}

func newDeferredExporter() *deferredExporter {
	return &deferredExporter{}
}

func (e *deferredExporter) Bind(delegate Exporter) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.delegate != nil {
		return nil
	}

	e.delegate = delegate
	var firstErr error
	for _, metric := range e.metrics {
		if err := metric.bind(delegate); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	e.metrics = nil
	return firstErr
}

func (e *deferredExporter) NewCounter(name, desc string, labelNames []string) Counter {
	counter := &deferredCounter{name: name, desc: desc, labelNames: labelNames}
	e.track(counter)
	return counter
}

func (e *deferredExporter) NewGauge(name, desc string, labelNames []string) Gauge {
	gauge := &deferredGauge{name: name, desc: desc, labelNames: labelNames}
	e.track(gauge)
	return gauge
}

func (e *deferredExporter) NewHistogram(name, desc string, buckets []float64, labelNames []string) Histogram {
	histogram := &deferredHistogram{name: name, desc: desc, buckets: buckets, labelNames: labelNames}
	e.track(histogram)
	return histogram
}

func (e *deferredExporter) track(metric bindableMetric) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.delegate == nil {
		e.metrics = append(e.metrics, metric)
		return
	}
	_ = metric.bind(e.delegate)
}

func (e *deferredExporter) HTTPHandler() http.Handler {
	e.mu.RLock()
	defer e.mu.RUnlock()
	if e.delegate == nil {
		return nil
	}
	return e.delegate.HTTPHandler()
}

type deferredMetric struct {
	mu                 sync.RWMutex
	delegate           Metric
	registrationQueued bool
}

func (m *deferredMetric) bind(delegate Metric) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.delegate != nil {
		return nil
	}
	m.delegate = delegate
	if m.registrationQueued {
		return delegate.Register()
	}
	return nil
}

func (m *deferredMetric) Register() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.delegate == nil {
		m.registrationQueued = true
		return nil
	}
	return m.delegate.Register()
}

func (m *deferredMetric) Unregister() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.delegate == nil {
		m.registrationQueued = false
		return false
	}
	return m.delegate.Unregister()
}

func (m *deferredMetric) Reset() {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.delegate != nil {
		m.delegate.Reset()
	}
}

func (m *deferredMetric) getDelegate() Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.delegate
}

type deferredCounter struct {
	deferredMetric
	name       string
	desc       string
	labelNames []string
}

func (c *deferredCounter) bind(exporter Exporter) error {
	return c.deferredMetric.bind(exporter.NewCounter(c.name, c.desc, c.labelNames))
}

func (c *deferredCounter) Add(value float64, labelValues ...string) {
	if delegate := c.getDelegate(); delegate != nil {
		delegate.(Counter).Add(value, labelValues...)
	}
}

type deferredGauge struct {
	deferredMetric
	name       string
	desc       string
	labelNames []string
}

func (g *deferredGauge) bind(exporter Exporter) error {
	return g.deferredMetric.bind(exporter.NewGauge(g.name, g.desc, g.labelNames))
}

func (g *deferredGauge) Set(value float64, labelValues ...string) {
	if delegate := g.getDelegate(); delegate != nil {
		delegate.(Gauge).Set(value, labelValues...)
	}
}

type deferredHistogram struct {
	deferredMetric
	name       string
	desc       string
	buckets    []float64
	labelNames []string
}

func (h *deferredHistogram) bind(exporter Exporter) error {
	return h.deferredMetric.bind(exporter.NewHistogram(h.name, h.desc, h.buckets, h.labelNames))
}

func (h *deferredHistogram) Observe(value float64, labelValues ...string) {
	if delegate := h.getDelegate(); delegate != nil {
		delegate.(Histogram).Observe(value, labelValues...)
	}
}
