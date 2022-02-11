// Copyright 1999-2021 Alibaba Group Holding Ltd.
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
	"os"
	"strconv"

	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/exporter/metric/prometheus"
)

var (
	exporter Exporter

	host      string
	app       string
	pid       string
	namespace string
)

func init() {
	if config.MetricExportHTTPAddr() != "" {
		exporter = newPrometheusExporter()
	} else {
		exporter = newEmptyExporter()
	}

	host, _ = os.Hostname()
	if host == "" {
		host = "unknown"
	}
	app = config.AppName()
	pid = strconv.Itoa(os.Getpid())
	namespace = "sentinel_go"
}

// Metric models basic operations of metric being exported.
// Implementations of Metric in this package are Counter, Gauge and Histogram.
type Metric interface {
	// Register registers the Metric.
	Register() error
	// Unregister unregisters the Metric.
	Unregister() bool
	// Reset deletes all values of the Metric.
	Reset()
}

// Counter is a Metric that represents a single numerical value that only ever goes up.
type Counter interface {
	Metric
	Add(value float64, labelValues ...string)
}

// Gauge is a Metric that represents a single numerical value that can arbitrarily go up and down.
type Gauge interface {
	Metric
	Set(value float64, labelValues ...string)
}

// Histogram counts individual observations from an event or sample stream in configurable buckets.
type Histogram interface {
	Metric
	Observe(value float64, labelValues ...string)
}

// Exporter creates all kinds Metric, and return the http.Handler used to export metrics.
type Exporter interface {
	// NewCounter creates a Counter metric partitioned by the given label names.
	NewCounter(name, desc string, labelNames []string) Counter
	// NewGauge creates a Gauge metric partitioned by the given label names.
	NewGauge(name, desc string, labelNames []string) Gauge
	// NewHistogram creates a histogram metric partitioned by the given label names.
	NewHistogram(name, desc string, buckets []float64, labelNames []string) Histogram
	// HTTPHandler returns http.Handler used to export metrics.
	HTTPHandler() http.Handler
}

// prometheusExporter is the exporter of prometheus implementation.
type prometheusExporter struct{}

func newPrometheusExporter() *prometheusExporter {
	return &prometheusExporter{}
}

func (e *prometheusExporter) NewCounter(name, desc string, labelNames []string) Counter {
	return prometheus.NewCounter(name, namespace, desc, labelNames, newConstLabels())
}

func (e *prometheusExporter) NewGauge(name, desc string, labelNames []string) Gauge {
	return prometheus.NewGauge(name, namespace, desc, labelNames, newConstLabels())
}

func (e *prometheusExporter) NewHistogram(name, desc string, buckets []float64, labelNames []string) Histogram {
	return prometheus.NewHistogram(name, namespace, desc, buckets, labelNames, newConstLabels())
}

func (e *prometheusExporter) HTTPHandler() http.Handler {
	return prometheus.HTTPHandler()
}

func newConstLabels() map[string]string {
	return map[string]string{
		"host": host,
		"app":  app,
		"pid":  pid,
	}
}

// NewCounter creates a Counter metric partitioned by the given label names.
func NewCounter(name, desc string, labelNames []string) Counter {
	return exporter.NewCounter(name, desc, labelNames)
}

// NewGauge creates a Gauge metric partitioned by the given label names.
func NewGauge(name, desc string, labelNames []string) Gauge {
	return exporter.NewGauge(name, desc, labelNames)
}

// NewHistogram creates a histogram metric partitioned by the given label names.
func NewHistogram(name, desc string, buckets []float64, labelNames []string) Histogram {
	return exporter.NewHistogram(name, desc, buckets, labelNames)
}

// HTTPHandler returns http.Handler used to export metrics.
func HTTPHandler() http.Handler {
	return exporter.HTTPHandler()
}

// Register registers the provided Metric.
func Register(m Metric) error {
	return m.Register()
}

// MustRegister registers the provided Metric and panics if any error occurs.
func MustRegister(m Metric) {
	if err := m.Register(); err != nil {
		panic(err)
	}
}
