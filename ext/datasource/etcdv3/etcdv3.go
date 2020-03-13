package etcdv3

import (
	"context"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"time"
)

var logger = logging.GetDefaultLogger()

type etcdv3DataSource struct{
	handler []datasource.PropertyHandler
	client *clientv3.Client
	ruleKey string
}

func (c *etcdv3DataSource)AddPropertyHandler(h datasource.PropertyHandler){
	c.handler = append(c.handler, h)
}

func (c *etcdv3DataSource)RemovePropertyHandler(h datasource.PropertyHandler){
	for index, value := range c.handler{
		if value == h{
			c.handler = append(c.handler[:index], c.handler[index+1:]...)
			break
		}
	}
}

func (c *etcdv3DataSource)ReadSource() ([]byte, error){
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.client.Get(ctx, c.ruleKey)
	if err != nil{
		return nil, err
	}
	if resp.Count == 0{
		return nil, errors.Errorf("No %v key in etcd", c.ruleKey)
	}
	return resp.Kvs[0].Value, nil
}

func (c *etcdv3DataSource)Initialize() error{
	go util.RunWithRecover(c.watch, logger)
	newValue, err := c.ReadSource()
	if err != nil{
		logger.Warnf("[EtcdDataSource] Initial configuration is null, you may have to check your data source")
	}else{
		c.updateValue(newValue)
	}
	return err
}

func (c *etcdv3DataSource)updateValue(newValue []byte){
	for _, handler := range c.handler{
		err := handler.Handle(newValue)
		if err != nil{
			logger.Warnf("Handler:%+v update property failed with error: %+v", handler, err)
		}
	}
}

func (c *etcdv3DataSource)watch(){
	for{
		rch := c.client.Watch(context.Background(), c.ruleKey)
		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT{
					newValue, err := c.ReadSource()
					if err != nil{
						logger.Warnf("receive etcd put event, but fail to read configuration from etcd")
						continue
					}
					c.updateValue(newValue)
				}
				if ev.Type == mvccpb.DELETE{
					c.updateValue(nil)
				}
			}
		}
	}
}
func (c *etcdv3DataSource)Close() error {
	if c.client != nil{
		err := c.client.Close()
		return err
	}
	return nil
}
func NewEtcdDataSource(key string, handler ...datasource.PropertyHandler) (*etcdv3DataSource, error){
	var err error
	ds := &etcdv3DataSource{
		handler: handler,
		client:  nil,
		ruleKey: "",
	}
	ds.ruleKey = key
	if !isAuthEnable() {
		ds.client, err = clientv3.New(clientv3.Config{Endpoints:getEndPoint()})
	} else {
		ds.client, err = clientv3.New(clientv3.Config{Endpoints:getEndPoint(), Username:getUser(),Password:getPassWord()})
	}
	if err != nil{
		logger.Errorf("Etcd client init failed with error: %+v", err)
		return nil, err
	}
	err = ds.Initialize()
	return ds, err
}