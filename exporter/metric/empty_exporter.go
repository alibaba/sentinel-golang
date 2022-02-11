package metric

import "net/http"

type EmptyExporter struct {
}

func (e *EmptyExporter) HTTPHandler() http.Handler {
	return nil
}

type emptyMetric struct {
}

func (e emptyMetric) Register() error {
	return nil
}

func (e emptyMetric) Unregister() bool {
	return false
}

func (e emptyMetric) Reset() {
	return
}

func newEmptyExporter() *EmptyExporter {
	return &EmptyExporter{}
}

type emptyCounter struct {
	emptyMetric
}

func (e emptyCounter) Add(value float64, labelValues ...string) {
	return
}

func (e *EmptyExporter) NewCounter(name, desc string, labelNames []string) Counter {
	return emptyCounter{}
}

type emptyGauge struct {
	emptyMetric
}

func (e emptyGauge) Set(value float64, labelValues ...string) {
	return
}

func (e *EmptyExporter) NewGauge(name, desc string, labelNames []string) Gauge {
	return &emptyGauge{}
}

type emptyHistogram struct {
	emptyMetric
}

func (e emptyHistogram) Observe(value float64, labelValues ...string) {
	return
}

func (e *EmptyExporter) NewHistogram(name, desc string, buckets []float64, labelNames []string) Histogram {
	return &emptyHistogram{}
}
