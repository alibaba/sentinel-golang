package apollo

import (
	"reflect"
	"testing"

	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/shima-park/agollo"
	"github.com/stretchr/testify/assert"
)

const (
	TestSystemRulesV1 = `[
    {
        "id": "0",
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": "1",
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": "2",
        "metricType": 0,
        "adaptiveStrategy": 0
    }
]`

	TestSystemRulesV2 = `[
    {
        "id": "0",
        "metricType": 0,
        "adaptiveStrategy": 0
    }
]`

	TestNamespace = "system-rules.json"
)

var (
	SystemRulesV1 = []*system.Rule{
		{ID: "0", MetricType: 0, Strategy: 0},
		{ID: "1", MetricType: 0, Strategy: 0},
		{ID: "2", MetricType: 0, Strategy: 0},
	}

	SystemRulesV2 = []*system.Rule{
		{ID: "0", MetricType: 0, Strategy: 0},
	}
)

type mockApollo struct {
	watchCh    chan *agollo.ApolloResponse
	namespaces map[string]agollo.Configurations
}

func newMockApollo() *mockApollo {
	return &mockApollo{
		watchCh: make(chan *agollo.ApolloResponse),
		namespaces: map[string]agollo.Configurations{
			TestNamespace: {
				"content": TestSystemRulesV1,
			},
		},
	}
}

func (a *mockApollo) Start() <-chan *agollo.LongPollerError {
	return make(chan *agollo.LongPollerError)
}

func (a *mockApollo) Stop() {}

func (a *mockApollo) Get(key string, opts ...agollo.GetOption) string {
	return ""
}

func (a *mockApollo) GetNameSpace(namespace string) agollo.Configurations {
	return a.namespaces[namespace]
}

func (a *mockApollo) Watch() <-chan *agollo.ApolloResponse {
	return a.watchCh
}

func (a *mockApollo) WatchNamespace(namespace string, stop chan bool) <-chan *agollo.ApolloResponse {
	return a.watchCh
}

func (a *mockApollo) Options() agollo.Options {
	return agollo.Options{}
}

func (a *mockApollo) publish(content string) {
	newValue := agollo.Configurations{"content": content}
	a.namespaces[TestNamespace] = newValue
	a.watchCh <- &agollo.ApolloResponse{NewValue: newValue}
}

func TestApolloDatasource(t *testing.T) {
	mock := newMockApollo()
	ds, err := NewDatasource(mock, TestNamespace)
	assert.Nil(t, err)

	expectRules := SystemRulesV1
	ds.AddPropertyHandler(datasource.NewDefaultPropertyHandler(
		datasource.SystemRuleJsonArrayParser,
		func(rule interface{}) error {
			assert.True(t, reflect.DeepEqual(expectRules, rule))
			return nil
		},
	))

	assert.Nil(t, ds.Initialize())
	assert.EqualError(t, ds.Initialize(), "Apollo datasource had been initialized")

	t.Run("WatchNamespaceChange", func(t *testing.T) {
		expectRules = SystemRulesV2
		// publish changes
		mock.publish(TestSystemRulesV2)
	})
}
