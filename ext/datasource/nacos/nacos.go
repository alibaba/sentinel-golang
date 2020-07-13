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

type ConfigServerInfo struct {
	Ip          string
	Port        uint64
	Username    string
	Password    string
	DataId      string
	Group       string
	NamespaceId string
}

var clientConfig = constant.ClientConfig{
	TimeoutMs:           10 * 1000,
	BeatInterval:        5 * 1000,
	ListenInterval:      300 * 1000,
	NotLoadCacheAtStart: true,
}

type NacosDataSource struct {
	datasource.Base
	configServerInfo ConfigServerInfo
	configClient     *config_client.ConfigClient
	isInitialized    util.AtomicBool
	configParam      *vo.ConfigParam
}

func NewNacosDataSource(configServerInfo ConfigServerInfo, handlers ...datasource.PropertyHandler) *NacosDataSource {
	var ds = &NacosDataSource{
		configServerInfo: configServerInfo,
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds
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
	s.configClient = &client
	err = s.listen(s.configClient)
	return err
}

func buildNacosClient(s *NacosDataSource) (nacos_client.INacosClient, error) {
	nc := nacos_client.NacosClient{}
	nc.SetServerConfig([]constant.ServerConfig{constant.ServerConfig{
		IpAddr:      s.configServerInfo.Ip,
		Port:        s.configServerInfo.Port,
		ContextPath: "/nacos",
	}})
	clientConfig.Password = s.configServerInfo.Password
	clientConfig.Username = s.configServerInfo.Username
	clientConfig.NamespaceId = s.configServerInfo.NamespaceId
	err := nc.SetClientConfig(clientConfig)
	err = nc.SetHttpAgent(&http_agent.HttpAgent{})
	return &nc, err
}
func (s *NacosDataSource) ReadSource() ([]byte, error) {
	content, err := s.configClient.GetConfig(vo.ConfigParam{
		DataId: s.configServerInfo.DataId,
		Group:  s.configServerInfo.Group,
	})
	if err != nil {
		return nil, errors.Errorf("Failed to read the nacos data source, err: %+v", err)
	}
	return []byte(content), err
}

func (s *NacosDataSource) listen(client *config_client.ConfigClient) (err error) {
	s.configParam = &vo.ConfigParam{
		DataId: s.configServerInfo.DataId,
		Group:  s.configServerInfo.Group,
		OnChange: func(namespace, group, dataId, data string) {
			logger.Infof("Configuration update, data content:[%s]", data)
			s.Handle([]byte(data))
		},
		ListenCloseChan: make(chan struct{},1),
	}
	err = client.ListenConfig(*s.configParam)
	if err != nil {
		return errors.Errorf("Failed to listen to the nacos data source, err: %+v", err)
	}
	return
}

func (s *NacosDataSource) Close() error {
	s.configParam.ListenCloseChan <- struct{}{}
	logger.Infof("The RefreshableFileDataSource   had been closed. DataId:[%s],Group:[%s]",
		s.configServerInfo.DataId, s.configServerInfo.Group)
	return nil
}
