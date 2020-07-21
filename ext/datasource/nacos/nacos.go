package nacos

import (
	"fmt"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
)

var (
	logger = logging.GetDefaultLogger()
)

type NacosDataSource struct {
	datasource.Base
	client         config_client.IConfigClient
	isInitialized  util.AtomicBool
	propertyConfig vo.ConfigParam
	closeChan      chan struct{}
}

func NewNacosDataSource(client config_client.IConfigClient, propertyConfig vo.ConfigParam, handlers ...datasource.PropertyHandler) (*NacosDataSource, error) {
	if client == nil {
		return nil, errors.New("Nil nacos config client")
	}
	if len(propertyConfig.Group) == 0 || len(propertyConfig.DataId) == 0 {
		return nil, errors.New(fmt.Sprintf("Invalid property config, group: %s, dataId: %s.", propertyConfig.Group, propertyConfig.DataId))
	}
	var ds = &NacosDataSource{
		Base:           datasource.Base{},
		client:         client,
		propertyConfig: propertyConfig,
		closeChan:      make(chan struct{}, 1),
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds, nil
}

func (s *NacosDataSource) Initialize() error {
	if !s.isInitialized.CompareAndSet(false, true) {
		return nil
	}
	data, err := s.ReadSource()
	if err != nil {
		return err
	}
	if err = s.doUpdate(data); err != nil {
		return err
	}
	err = s.listen(s.client)
	if err == nil {
		logger.Infof("Nacos data source is successfully initialized, property config: %+v", s.propertyConfig)
	}
	return err
}

func (s *NacosDataSource) ReadSource() ([]byte, error) {
	content, err := s.client.GetConfig(s.propertyConfig)
	if err != nil {
		return nil, errors.Errorf("Failed to read the nacos data source when initialization, err: %+v", err)
	}

	logger.Infof("Succeed to read source for property config: %+v, data: %s", s.propertyConfig, content)
	return []byte(content), err
}

func (s *NacosDataSource) doUpdate(data []byte) error {
	return s.Handle(data)
}

func (s *NacosDataSource) listen(client config_client.IConfigClient) (err error) {
	listener := vo.ConfigParam{
		DataId: s.propertyConfig.DataId,
		Group:  s.propertyConfig.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logger.Infof("Receive listened property. namespace: %s, group: %s, dataId: %s, data: %s.", namespace, group, dataId, data)
			err := s.doUpdate([]byte(data))
			if err != nil {
				logger.Errorf("Fail to update data source, err: %+v", err)
			}
		},
		ListenCloseChan: s.closeChan,
	}
	err = client.ListenConfig(listener)
	if err != nil {
		return errors.Errorf("Failed to listen to the nacos data source, err: %+v", err)
	}
	return nil
}

func (s *NacosDataSource) Close() error {
	s.closeChan <- struct{}{}
	logger.Infof("The nacos datasource had been closed. property config: %+v", s.propertyConfig)
	return nil
}
