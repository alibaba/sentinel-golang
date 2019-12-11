package etcdv3

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sentinel-group/sentinel-golang/core"
)

type DataSource struct {
	*core.DynamicSentinelProperty
	client *clientv3.Client
	ruleKey string
	watcher *Watch
}

// New ...
func New(client *clientv3.Client, ruleKey string) (*DataSource, error) {
	ds := &DataSource{
		DynamicSentinelProperty: core.NewDynamicSentinelProperty(),
		client: client,
		ruleKey: ruleKey,
		watcher: newWatch(client, ruleKey),
	}

	if err := ds.loadConfig(); err != nil {
		return nil, err
	}

	go ds.watch()

	return ds, nil
}

func (ds *DataSource) watch() {
	for event := range ds.watcher.C() {
		switch event.Type {
		case mvccpb.PUT:
			if err := ds.UpdateValue(event.Kv.Value); err != nil {
				log.Printf("property update value err: %+v\n", err)
			}
		case mvccpb.DELETE:
			if err := ds.UpdateValue([]byte{'[',']'}); err != nil {
				log.Printf("property delete value err: %+v\n", err)
			}
		}
	}
}

func (ds *DataSource) loadConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := ds.client.Get(ctx, ds.ruleKey)
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		if err := ds.SetValue(kv.Value); err != nil {
			return err
		}
	}

	return nil
}

func (ds *DataSource) Close() error {
	ds.watcher.Close()
	return nil
}
