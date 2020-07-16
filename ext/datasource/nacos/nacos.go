package nacos

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/pkg/errors"
)

var (
	logger = logging.GetDefaultLogger()
)

type ConfigParam struct {
	DataId string
	Group  string
}

var clientConfig = constant.ClientConfig{
	TimeoutMs:      10 * 1000,
	ListenInterval: 30 * 1000,
	BeatInterval:   5 * 1000,
	LogDir:         "/nacos/logs",
	CacheDir:       "/nacos/cache",
}

type NacosDataSource struct {
	datasource.Base
	serverConfig  constant.ServerConfig
	clientConfig  constant.ClientConfig
	client        *config_client.ConfigClient
	isInitialized util.AtomicBool
	configParam   ConfigParam
	listener      *vo.ConfigParam
}

func NewNacosDataSource(clientConfig constant.ClientConfig, serverConfig constant.ServerConfig, configParam ConfigParam, handlers ...datasource.PropertyHandler) (*NacosDataSource, error) {
	if len(serverConfig.IpAddr) <= 0 || serverConfig.Port <= 0 {
		return nil, errors.New("The nacos serverConfig is incorrect.")
	}
	var ds = &NacosDataSource{
		clientConfig: clientConfig,
		serverConfig: serverConfig,
		configParam:  configParam,
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
	nc, err := buildNacosClient(s)
	client, err := config_client.NewConfigClient(nc)
	if err != nil {
		return errors.Errorf("Nacosclient failed to build, err: %+v", err)
	}
	s.client = &client
	err = s.listen(s.client)
	return err
}

func buildNacosClient(s *NacosDataSource) (nacos_client.INacosClient, error) {
	nc := nacos_client.NacosClient{}
	nc.SetServerConfig([]constant.ServerConfig{s.serverConfig})
	err := nc.SetClientConfig(s.clientConfig)
	err = nc.SetHttpAgent(&http_agent.HttpAgent{})
	return &nc, err
}
func (s *NacosDataSource) ReadSource() ([]byte, error) {
	content, err := s.client.GetConfig(vo.ConfigParam{
		DataId: s.configParam.DataId,
		Group:  s.configParam.Group,
	})
	if err != nil {
		return nil, errors.Errorf("Failed to read the nacos data source, err: %+v", err)
	}
	return []byte(content), err
}

func (s *NacosDataSource) listen(client *config_client.ConfigClient) (err error) {
	s.listener = &vo.ConfigParam{
		DataId: s.configParam.DataId,
		Group:  s.configParam.Group,
		OnChange: func(namespace, group, dataId, data string) {
			if err := s.Handle([]byte(data)); err != nil {
				logger.Errorf("Fail to update data for dataId:[%s] group:[%s] namespaceId:[%s]  when execute "+
					"listen function, err: %+v", data, group, namespace, err)
			}

		},
		ListenCloseChan: make(chan struct{}, 1),
	}
	err = client.ListenConfig(*s.listener)
	if err != nil {
		return errors.Errorf("Failed to listen to the nacos data source, err: %+v", err)
	}
	return
}

func (s *NacosDataSource) Close() error {
	s.listener.ListenCloseChan <- struct{}{}

	logger.Infof("The RefreshableFileDataSource   had been closed. DataId:[%s],Group:[%s]",
		s.configParam.DataId, s.configParam.Group)
	return nil
}
