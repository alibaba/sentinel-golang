package etcdv3

import (
	"fmt"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/stretchr/testify/assert"
)

var tl = testListener{}
type testListener struct { }

func (tp testListener) ConfigLoad(data []byte) error {
	fmt.Printf("load string(data) = %+v\n", string(data))
	return nil
}
func (tp testListener) ConfigUpdate(data []byte) error {
	fmt.Printf("update string(data) = %+v\n", string(data))
	return nil
}

func TestDataSource_LoadConfig(t *testing.T) {
	config := clientv3.Config{
			Endpoints:            []string{"etcd-naming.dz11.com:2379"},
			DialTimeout:          time.Second,
			DialKeepAliveTime:    time.Second,
			DialKeepAliveTimeout: time.Second,
	}

	client, err := clientv3.New(config)
	assert.Nil(t, err)

	ds, err := New(client, "/wsd-sentinel/appname/flow")
	assert.Nil(t, err)
	assert.NotNil(t, ds)
	ds.AddListener(tl)
	// ds.AddListener(flow.NewRuleManager())
	assert.Nil(t, ds.loadConfig())

	time.Sleep(time.Second * 100)
}
