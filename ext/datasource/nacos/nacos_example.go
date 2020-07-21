package nacos

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

var serverConfig = constant.ServerConfig{
	ContextPath: "/nacos",
	Port:        8848,
	IpAddr:      "127.0.0.1",
}

var clientConfigTest = constant.ClientConfig{
	BeatInterval:   10000,
	TimeoutMs:      10000,
	ListenInterval: 20000,
}

func cretateConfigClientTest() (*config_client.ConfigClient, error) {
	nc := nacos_client.NacosClient{}
	err := nc.SetServerConfig([]constant.ServerConfig{serverConfig})
	err = nc.SetClientConfig(clientConfigTest)
	err = nc.SetHttpAgent(&http_agent.HttpAgent{})
	client, err := config_client.NewConfigClient(&nc)

	return &client, err
}

func Example_NacosDatasource_CustomizeClient() {
	client, err := cretateConfigClientTest()
	if err != nil {
		// todo something
	}
	nds, err := NewNacosDataSource(client, vo.ConfigParam{
		DataId: "system-rules",
		Group:  "sentinel-go",
	}, &datasource.MockPropertyHandler{})
	if err != nil {
		// todo something
	}
	err = nds.Initialize()
	if err != nil {
		// todo something
	}
}
