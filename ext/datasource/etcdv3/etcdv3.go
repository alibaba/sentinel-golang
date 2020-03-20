package etcdv3

import (
	"context"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"time"
)

var logger = logging.GetDefaultLogger()

type Etcdv3DataSource struct{
	datasource.Base
	client *clientv3.Client
	propertyKey string
	watchIndex	int64
}

func (c *Etcdv3DataSource)ReadSource() ([]byte, error){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.client.Get(ctx, c.propertyKey)
	if err != nil{
		return nil, errors.Errorf("The key[%s] is not existed in etcd.", c.propertyKey)
	}
	c.watchIndex = resp.Header.Revision
	return resp.Kvs[0].Value, nil
}

func (c *Etcdv3DataSource)Initialize() error{
	err := c.doReadAndUpdate()
	if err != nil {
		logger.Errorf("[Etcdv3DataSource]Fail to execute doReadAndUpdate, err: %+v", err)
	}
	go util.RunWithRecover(c.watch, logger)
	return err
}

func (c *Etcdv3DataSource) doReadAndUpdate() error {
	src, err := c.ReadSource()
	if err != nil {
		err = errors.Errorf("[Etcdv3DataSource]Fail to read source, err: %+v", err)
		return err
	}
	for _, h := range c.Handlers() {
		e := h.Handle(src)
		if e != nil {
			err = multierr.Append(err, e)
		}
	}
	return err
}

func (c *Etcdv3DataSource)watch(){
	for{
		rch := c.client.Watch(context.Background(), c.propertyKey, clientv3.WithRev(int64(c.watchIndex)))
		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT{
					err := c.doReadAndUpdate()
					if err != nil{
						logger.Errorf("Fail to execute doReadAndUpdate for PUT event, err: %+v", err)
					}
				}
				if ev.Type == mvccpb.DELETE{
					var updateErr error
					for _, h := range c.Handlers() {
						e := h.Handle(nil)
						if e != nil {
							updateErr = multierr.Append(updateErr, e)
						}
					}
					if updateErr != nil{
						logger.Errorf("Fail to execute doReadAndUpdate for DELETE event, err: %+v", updateErr)
					}
				}
			}
		}
	}
}

func (c *Etcdv3DataSource)Close() error {
	if c.client != nil{
		err := c.client.Close()
		return err
	}
	return nil
}
func NewEtcdv3DataSource(key string, cfg clientv3.Config, handlers ...datasource.PropertyHandler) (*Etcdv3DataSource, error){
	var err error
	ds := &Etcdv3DataSource{
		client:  nil,
		propertyKey: "",
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	ds.propertyKey = key
	ds.client, err = clientv3.New(cfg)
	if err != nil{
		logger.Errorf("Fail to create etcd client, err: %+v", err)
		return nil, err
	}
	err = ds.Initialize()
	return ds, err
}