/*
Before you run this demo, you should install etcd in your local machine or you can
change 127.0.0.1:2379 to your machine which have etcd cluster.
*/
package main

import (
	"context"
	"encoding/json"
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/ext/datasource/etcdv3"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"log"
	"math/rand"
	"time"
)

func WriteDataToLocalEtcd() {
	client, err := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}, DialTimeout: time.Second})
	data := []*flow.FlowRule{
		{
			Resource:        "some-test",
			MetricType:      flow.QPS,
			Count:           1000,
			ControlBehavior: flow.Reject,
		},
		{
			Resource:        "some-test",
			MetricType:      flow.QPS,
			Count:           10,
			ControlBehavior: flow.Reject,
		},
	}
	log.Println(client, err)
	if err != nil {
		log.Fatalf("Fail to create etcd client, err: %+v", err)
		return
	}
	value, _ := json.Marshal(data)
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	_, err = client.Put(ctx, "flow", string(value))
	if err != nil {
		log.Fatalf("Fail to put data into etcd, err: %+v", err)
	}
}

func Delete(client *clientv3.Client) {
	client.Delete(context.Background(), "flow")
}

// The function will update etcd data every two second.
func OperationEtcd(client *clientv3.Client) {
	t1 := time.NewTimer(2 * time.Second)
	flag := 0
	for {
		select {
		case <-t1.C:
			if flag == 0 {
				Delete(client)
				flag = 1
			} else {
				WriteDataToLocalEtcd()
				flag = 0
			}
			t1.Reset(time.Second * 2)
		}
	}
}
func main() {
	//Write the default configuration into etcd.
	WriteDataToLocalEtcd()
	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	var dataSourceClient datasource.DataSource
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	cfg := &clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second,
	}
	handler := datasource.NewDefaultPropertyHandler(datasource.FlowRulesJsonConverter, datasource.FlowRulesUpdater)
	dataSourceClient, err = etcdv3.NewEtcdv3DataSource("flow", cfg, handler)
	if err != nil {
		log.Fatalf("Fail to create etcd client, err: %+v", err)
		return
	}
	err = dataSourceClient.Initialize()
	if err != nil {
		log.Fatalf("Load initialization from etcd failed, err: %+v", err)
		return
	}
	defer dataSourceClient.Close()
	client, _ := clientv3.New(clientv3.Config{Endpoints: []string{"127.0.0.1:2379"}})
	go OperationEtcd(client)
	ch := make(chan struct{})

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
	<-ch
}
