package etcdv3

import (
	"context"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/extension/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"go.uber.org/multierr"
)

type etcdv3DataSource struct {
	datasource.Base
	client *clientv3.Client
	ruleKey string
	watcher *Watch
	logger logging.Logger
}

// New create new etcdv3DataSource instance
// todo(gorexlv): add option builder to support
func New(client *clientv3.Client, ruleKey string) base.DataSource {
	ds := &etcdv3DataSource{
		Base: datasource.Base{DataFormat:"json"},
		client: client,
		ruleKey: ruleKey,
		watcher: newWatch(client, ruleKey),
		logger: logging.GetDefaultLogger(),
	}

	return ds
}

// ReadConfig implements base.DataSource interface
func (ds *etcdv3DataSource) ReadConfig() error {
	if err := ds.loadConfig(); err != nil {
		logging.GetDefaultLogger().Error("load config", "err", err.Error())
		return err
	}

	go ds.watch()
	return nil
}

// Close implements base.DataSource interface
func (ds *etcdv3DataSource) Close() error {
	ds.watcher.Close()
	return nil
}

func (ds *etcdv3DataSource) watch() {
	for event := range ds.watcher.C() {
		switch event.Type {
		// todo(gorexlv): handler for each event type?
		case mvccpb.PUT:
			if err := ds.ApplyConfig(event.Kv.Value); err != nil {
				logging.GetDefaultLogger().Error("put config", "bytes", string(event.Kv.Value), "err", err.Error())
			}
		case mvccpb.DELETE:
			if err := ds.DeleteConfig(); err != nil {
				logging.GetDefaultLogger().Error("delete config", "bytes", string(event.Kv.Value), "err", err.Error())
			}
		}
	}
}

func (ds *etcdv3DataSource) loadConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := ds.client.Get(ctx, ds.ruleKey)
	if err != nil {
		return err
	}

	for _, kv := range resp.Kvs {
		err = multierr.Append(err, ds.ApplyConfig(kv.Value))
	}

	return err
}


