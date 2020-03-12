package etcdv3

import (
	"context"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
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

func (c *etcdv3DataSource)ReadSource() []byte{
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, _ := c.client.Get(ctx, c.ruleKey)
	if resp.Count == 0{
		logger.Warnf("Get data from etcd failed")
		return nil
	}
	return resp.Kvs[0].Value
}

func (c *etcdv3DataSource)Initialize(){
	go util.RunWithRecover(c.watch, logger)
	newValue := c.ReadSource()
	if newValue != nil{
		c.updateValue(newValue)
	}else{
		logger.Warnf("[EtcdDataSource] Initial configuration is null, you may have to check your data source")
	}
}

func (c *etcdv3DataSource)updateValue(newValue []byte){
	for _, handler := range c.handler{
		err := handler.Handle(newValue)
		logger.Warnf("Handler:%v update property failed with error: %v", handler, err)
	}
}

func (c *etcdv3DataSource)watch(){
	for{
		rch := c.client.Watch(context.Background(), c.ruleKey)
		for wresp := range rch {
			for _, ev := range wresp.Events {
				if ev.Type == mvccpb.PUT{
					newValue := c.ReadSource()
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
		logger.Errorf("Etcd client init failed with error: %v", err)
		return nil, err
	}
	ds.Initialize()
	return ds, err
}