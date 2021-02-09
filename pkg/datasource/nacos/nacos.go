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

type NacosDataSource struct {
	datasource.Base
	client        config_client.IConfigClient
	isInitialized util.AtomicBool
	group         string
	dataId        string
}

func NewNacosDataSource(client config_client.IConfigClient, group, dataId string, handlers ...datasource.PropertyHandler) (*NacosDataSource, error) {
	if client == nil {
		return nil, errors.New("nil nacos config client")
	}
	if len(group) == 0 || len(dataId) == 0 {
		return nil, errors.New(fmt.Sprintf("invalid parameters, group: %s, dataId: %s", group, dataId))
	}
	var ds = &NacosDataSource{
		Base:   datasource.Base{},
		client: client,
		group:  group,
		dataId: dataId,
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
		logging.Info("[Nacos] Nacos data source is successfully initialized", "group", s.group, "dataId", s.dataId)
	}
	return err
}

func (s *NacosDataSource) ReadSource() ([]byte, error) {
	content, err := s.client.GetConfig(vo.ConfigParam{
		DataId: s.dataId,
		Group:  s.group,
	})
	if err != nil {
		return nil, errors.Errorf("Failed to read the nacos data source when initialization, err: %+v", err)
	}

	logging.Info("[Nacos] Succeed to read source", "group", s.group, "dataId", s.dataId, "content", content)
	return []byte(content), err
}

func (s *NacosDataSource) doUpdate(data []byte) error {
	return s.Handle(data)
}

func (s *NacosDataSource) listen(client config_client.IConfigClient) (err error) {
	listener := vo.ConfigParam{
		DataId: s.dataId,
		Group:  s.group,
		OnChange: func(namespace, group, dataId, data string) {
			logging.Info("[Nacos] Receive listened property", "namespace", namespace, "group", group, "dataId", dataId, "data", data)
			err := s.doUpdate([]byte(data))
			if err != nil {
				logging.Error(err, "Fail to update data source in NacosDataSource.listen()")
			}
		},
	}
	err = client.ListenConfig(listener)
	if err != nil {
		return errors.Errorf("Failed to listen to the nacos data source, err: %+v", err)
	}
	return nil
}

func (s *NacosDataSource) Close() error {
	err := s.client.CancelListenConfig(vo.ConfigParam{
		DataId: s.dataId,
		Group:  s.group,
	})
	if err != nil {
		return errors.Errorf("Failed to cancel listen to the nacos data source, err: %+v", err)
	}
	logging.Info("[Nacos] The nacos datasource had been closed", "group", s.group, "dataId", s.dataId)
	return nil
}
