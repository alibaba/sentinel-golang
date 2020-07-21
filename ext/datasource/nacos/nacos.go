package nacos

import (
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
	client        config_client.IConfigClient
	isInitialized util.AtomicBool
	getConfig     vo.ConfigParam
	closeChan     chan struct{}
}

func NewNacosDataSource(client config_client.IConfigClient, getConfig vo.ConfigParam, handlers ...datasource.PropertyHandler) (*NacosDataSource, error) {
	var ds = &NacosDataSource{
		Base:      datasource.Base{},
		client:    client,
		getConfig: getConfig,
		closeChan: make(chan struct{}, 1),
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
	return s.listen(s.client)
}

func (s *NacosDataSource) ReadSource() ([]byte, error) {
	content, err := s.client.GetConfig(s.getConfig)
	if err != nil {
		return nil, errors.Errorf("Failed to read the nacos data source, err: %+v", err)
	}
	return []byte(content), err
}

func (s *NacosDataSource) doUpdate(data []byte) error {
	return s.Handle(data)
}

func (s *NacosDataSource) listen(client config_client.IConfigClient) (err error) {
	listener := vo.ConfigParam{
		DataId: s.getConfig.DataId,
		Group:  s.getConfig.Group,
		OnChange: func(namespace, group, dataId, data string) {
			err := s.doUpdate([]byte(data))
			if err != nil {
				logger.Errorf("")
			}
		},
		ListenCloseChan: s.closeChan,
	}
	err = client.ListenConfig(listener)
	if err != nil {
		return errors.Errorf("Failed to listen to the nacos data source, err: %+v", err)
	}
	return
}

func (s *NacosDataSource) Close() error {
	s.closeChan <- struct{}{}

	logger.Infof("The nacos datasource  had been closed. DataId:[%s],Group:[%s]",
		s.getConfig.DataId, s.getConfig.Group)
	return nil
}
