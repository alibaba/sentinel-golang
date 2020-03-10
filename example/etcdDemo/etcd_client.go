package main

import (
	"context"
	"encoding/json"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/coreos/etcd/clientv3"
)
func main(){
	client, _ := clientv3.New(clientv3.Config{Endpoints:[]string{"127.0.0.1:2379",}})
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
			Count:           0,
			ControlBehavior: flow.Reject,
		},
	}
	value, _ := json.Marshal(data)
	client.Put(context.Background(),"flow",string(value))
}
