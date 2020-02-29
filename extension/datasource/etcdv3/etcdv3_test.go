package etcdv3_test

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/extension/datasource/etcdv3"
	"github.com/coreos/etcd/clientv3"
	"gotest.tools/assert"
)

var testingClient *clientv3.Client
var testingExpect Data
var testingDataSource base.DataSource

type Data struct {
	A int `json:"a" toml:"a" yaml:"a"`
	B string `json:"b" toml:"b" yaml:"b"`
}

func TestMain(m *testing.M) {
	client, err := clientv3.New(clientv3.Config{
		// todo(gorexlv): add ENV variable to travis
		Endpoints:            strings.Split(os.Getenv("ETCD_ENDPOINTS"), ","),
		AutoSyncInterval:     0,
		DialTimeout:          time.Second * 5,
		DialKeepAliveTime:    0,
		DialKeepAliveTimeout: 0,
		MaxCallSendMsgSize:   0,
		MaxCallRecvMsgSize:   0,
		TLS:                  nil,
		RejectOldCluster:     false,
		DialOptions:          nil,
		Context:              nil,
	})
	if err != nil {
		panic(err)
	}
	testingClient = client
	testingDataSource = etcdv3.New(client, "test-etcd-ds")
	m.Run()
	testingDataSource.Close()
}

func initRegister(t *testing.T) {
	base.RegisterPropertyConsumer(func(decoder base.PropertyDecoder) error {
		var data Data
		err :=  decoder.Decode(&data)
		if err != nil && err != io.EOF {
			t.Fatalf("err: %+v\n", err.Error())
		}
		assert.Equal(t, data, testingExpect)
		return err
	}, func() error {
		return nil
	})
}

func updateKey(t *testing.T, raw []byte, expect Data) {
	testingExpect = expect
	resp, err := testingClient.Put(context.TODO(), "test-etcd-ds",string(raw) )
	if err != nil {
		t.Fatalf("put %+v\n", err)
	}
	t.Logf("resp: %+v\n", resp)
}

func deleteKey(t *testing.T) {
	testingExpect = Data{0,""}
	_, _ = testingClient.Delete(context.TODO(), "test-etcd-ds")
}

// ETCD_ENDPOINTS=127.0.0.1:2379 go test
func TestDatasource(t *testing.T) {
	initRegister(t)

	// open etcdv3DataSource
	updateKey(t, []byte(`{"A":1, "B":"2"}`), Data{A:1,B:"2"})
	assert.NilError(t, testingDataSource.ReadConfig())

	time.Sleep(time.Second)
	// watch key change
	updateKey(t, []byte(`{"A":3, "B":"4"}`), Data{A:3,B:"4"})
	time.Sleep(time.Second)

	// delete key
	deleteKey(t)
	time.Sleep(time.Second)
}

