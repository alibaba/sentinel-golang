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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recordingExporter struct {
	counter *recordingCounter
	handler http.Handler
}

func (e *recordingExporter) NewCounter(string, string, []string) Counter {
	e.counter = &recordingCounter{}
	return e.counter
}

func (e *recordingExporter) NewGauge(string, string, []string) Gauge {
	return &emptyGauge{}
}

func (e *recordingExporter) NewHistogram(string, string, []float64, []string) Histogram {
	return &emptyHistogram{}
}

func (e *recordingExporter) HTTPHandler() http.Handler {
	return e.handler
}

type recordingCounter struct {
	registered bool
	value      float64
}

func (c *recordingCounter) Register() error {
	c.registered = true
	return nil
}

func (c *recordingCounter) Unregister() bool {
	c.registered = false
	return true
}

func (c *recordingCounter) Reset() {
	c.value = 0
}

func (c *recordingCounter) Add(value float64, _ ...string) {
	c.value += value
}

func TestDeferredExporterBindsPredeclaredMetrics(t *testing.T) {
	exporter := newDeferredExporter()
	counter := exporter.NewCounter("requests", "request count", []string{"result"})
	require.NoError(t, counter.Register())

	// Observations before Sentinel initialization retain the old no-op behavior.
	counter.Add(1, "ok")
	recording := &recordingExporter{}
	require.NoError(t, exporter.Bind(recording))

	require.NotNil(t, recording.counter)
	assert.True(t, recording.counter.registered)
	counter.Add(2, "ok")
	assert.Equal(t, float64(2), recording.counter.value)
}

func TestDeferredExporterForwardsHTTPHandler(t *testing.T) {
	exporter := newDeferredExporter()
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	recording := &recordingExporter{handler: handler}
	require.NoError(t, exporter.Bind(recording))

	response := httptest.NewRecorder()
	exporter.HTTPHandler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	assert.Equal(t, http.StatusNoContent, response.Code)
}
