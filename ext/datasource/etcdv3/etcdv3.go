package etcdv3

import (
	"context"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"sync"
	"time"
)

var logger = logging.GetDefaultLogger()
var once sync.Once

type DatasourceGenerator struct {
	etcdv3Client *clientv3.Client
}

type Etcdv3DataSource struct {
	datasource.Base
	propertyKey         string
	lastUpdatedRevision int64
	client              *clientv3.Client
}

func (c *Etcdv3DataSource) ReadSource() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.client.Get(ctx, c.propertyKey)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, errors.Errorf("The key[%s] is not existed in etcd.", c.propertyKey)
	}
	c.lastUpdatedRevision = resp.Header.Revision
	return resp.Kvs[0].Value, nil
}

func (c *Etcdv3DataSource) Initialize() error {
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

func (c *Etcdv3DataSource) watch() {
	for {
		rch := c.client.Watch(context.Background(), c.propertyKey, clientv3.WithRev(int64(c.lastUpdatedRevision)))
		for wresp := range rch {
			err := wresp.Err()
			if err != nil {
				logger.Errorf("Fail to watch key[%s], err: %+v", c.propertyKey, err)
				if err == rpctypes.ErrCompacted {
					logger.Infof("Update lastUpdatedRevision:%v to CompactRevision:%v", c.lastUpdatedRevision, wresp.CompactRevision)
					c.lastUpdatedRevision = wresp.CompactRevision
				}
				continue
			}
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT {
					err := c.doReadAndUpdate()
					if err != nil {
						logger.Errorf("Fail to execute doReadAndUpdate for PUT event, err: %+v", err)
					}
				}
				if ev.Type == mvccpb.DELETE {
					var updateErr error
					for _, h := range c.Handlers() {
						e := h.Handle(nil)
						if e != nil {
							updateErr = multierr.Append(updateErr, e)
						}
					}
					if updateErr != nil {
						logger.Errorf("Fail to execute doReadAndUpdate for DELETE event, err: %+v", updateErr)
					}
				}
			}
		}
	}
}

func (c *Etcdv3DataSource) Close() error {
	if c.client != nil {
		err := c.client.Close()
		return err
	}
	return nil
}

func NewEtcdv3DatasourceGenerator(config *clientv3.Config) (*DatasourceGenerator, error) {
	var generator DatasourceGenerator
	var err error
	var client *clientv3.Client
	once.Do(func() {
		client, err = clientv3.New(*config)
		if err != nil {
			logger.Errorf("Fail to instance etcdv3 client, err: %+v", err)
		} else {
			generator.etcdv3Client = client
		}
	})
	return &generator, err
}

func (g *DatasourceGenerator) Generate(key string, handlers ...datasource.PropertyHandler) (*Etcdv3DataSource, error) {
	var err error
	if g.etcdv3Client == nil {
		err = errors.New("The etcdv3 client is nil in DatasourceGenerator")
		return nil, err
	}
	ds := &Etcdv3DataSource{
		propertyKey: key,
		client:      g.etcdv3Client,
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds, err
}
