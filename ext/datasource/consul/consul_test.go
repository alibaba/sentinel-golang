package consul

import (
	"sync"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/system"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	TestSystemRules = `[
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
)

var (
	SystemRules = []*system.Rule{
		{MetricType: 0, Strategy: 0},
		{MetricType: 0, Strategy: 0},
		{MetricType: 0, Strategy: 0},
	}
)

type consulClientMock struct {
	mock.Mock
	pair *api.KVPair
	lock sync.Mutex
}

func (c *consulClientMock) Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error) {
	c.lock.Lock()
	defer c.lock.Unlock()
	return c.pair, &api.QueryMeta{
		LastIndex:                 c.pair.ModifyIndex,
		LastContentHash:           "",
		LastContact:               0,
		KnownLeader:               false,
		RequestTime:               0,
		AddressTranslationEnabled: false,
		CacheHit:                  false,
		CacheAge:                  0,
	}, nil
}

func (c *consulClientMock) List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error) {
	panic("implement me")
}

func newQuerierMock() *consulClientMock {
	return &consulClientMock{
		pair: &api.KVPair{
			Key:         "property_key",
			CreateIndex: 0,
			ModifyIndex: 0,
			LockIndex:   0,
			Flags:       0,
			Value:       []byte(TestSystemRules),
			Session:     "",
		},
	}
}

func (c *consulClientMock) resetPair(pair *api.KVPair) {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.pair = pair
}

func TestConsulDatasource(t *testing.T) {
	mock := newQuerierMock()
	ds := newConsulDataSource("property_key", evaluateOptions([]Option{}))
	ds.kvQuerier = mock

	ds.AddPropertyHandler(datasource.NewDefaultPropertyHandler(
		datasource.SystemRuleJsonArrayParser,
		func(rule interface{}) error {
			assert.NotNil(t, rule)
			assert.ObjectsAreEqual(SystemRules, rule)
			return nil
		},
	))

	assert.Nil(t, ds.Initialize())
	assert.EqualError(t, ds.Initialize(), "consul datasource had been initialized")

	t.Run("WatchSourceChange", func(t *testing.T) {
		mock.resetPair(&api.KVPair{
			Key:         "property_key",
			CreateIndex: 0,
			ModifyIndex: 1,
			LockIndex:   0,
			Flags:       0,
			Value:       []byte(TestSystemRules),
			Session:     "",
		})
	})
	time.Sleep(time.Second)
}
