package etcdv3

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/sentinel-group/sentinel-golang/core/datasource"
	"github.com/sentinel-group/sentinel-golang/logging"
)

var logger = logging.GetDefaultLogger()

type DataSource struct {
	*datasource.SentinelProperty
	client *clientv3.Client
	ruleKey string
	watcher *Watch
	convert datasource.Converter
}

// New ...
func New(client *clientv3.Client, ruleKey string, convert datasource.Converter) (*DataSource, error) {
	ds := &DataSource{
		SentinelProperty: datasource.NewSentinelProperty(),
		client: client,
		ruleKey: ruleKey,
		watcher: newWatch(client, ruleKey),
		convert: convert,
	}

	if err := ds.loadConfig(); err != nil {
		return nil, err
	}

	go ds.watch()

	return ds, nil
}

func (ds *DataSource) watch() {
	for event := range ds.watcher.C() {
		fmt.Printf("event = %+v\n", event)
		switch event.Type {
		case mvccpb.PUT:
			val, err := ds.convert.Convert(event.Kv.Value)
			if err != nil {
				logger.Errorf("Error when converting data value: %+v", err)
			}

			if ok, err := ds.UpdateValue(val, datasource.FlagUpdate); err != nil {
				logger.Errorf("property update value err: %+v\n", err)
			} else if !ok {
				logger.Info("property update value failed")
			}
		case mvccpb.DELETE:
			if ok, err := ds.UpdateValue(nil, datasource.FlagDelete); err != nil {
				logger.Errorf("property delete value err: %+v\n", err)
			} else if !ok {
				logger.Info("property delete value failed")
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
		val, err := ds.convert.Convert(kv.Value)
		if err != nil {
			logger.Errorf("Error when converting data value: %+v", err)
		}
		if ok, err := ds.UpdateValue(val, datasource.FlagInitialLoaded); err != nil {
			return err
		} else if !ok {
			logger.Info("property load config failed")
		}
	}

	return nil
}

func (ds *DataSource) Close() error {
	ds.watcher.Close()
	return nil
}
