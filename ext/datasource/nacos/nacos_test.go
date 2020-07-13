package nacos

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
	"testing"
	"time"
)

const (
	TestSystemRules = `[
    {
        "id": 0,
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": 1,
        "metricType": 0,
        "adaptiveStrategy": 0
    },
    {
        "id": 2,
        "metricType": 0,
        "adaptiveStrategy": 0
    }
]`
)

var serverConfig = constant.ServerConfig{
	ContextPath: "/nacos",
	Port:        serverConfigTest.Port,
	IpAddr:      serverConfigTest.Ip,
}
var clientConfigTest = constant.ClientConfig{
	TimeoutMs:      10000,
	ListenInterval: 20000,
	BeatInterval:   10000,
}

var serverConfigTest = ConfigServerInfo{
	Port:   80,
	Ip:     "console.nacos.io",
	DataId: "system-rules",
	Group:  "sentinel-go",
}
var serverConfigErrTest = ConfigServerInfo{
	Port:   80,
	Ip:     "console.nacos.io1",
	DataId: "system-rules1",
	Group:  "sentinel-go1",
}

func cretateConfigClientTest() config_client.ConfigClient {
	nc := nacos_client.NacosClient{}
	nc.SetServerConfig([]constant.ServerConfig{serverConfig})
	nc.SetClientConfig(clientConfigTest)
	nc.SetHttpAgent(&http_agent.HttpAgent{})
	client, _ := config_client.NewConfigClient(&nc)

	return client
}

func prePushSystemRules(content string) (bool, error) {
	client := cretateConfigClientTest()
	success, err := client.PublishConfig(vo.ConfigParam{
		DataId:  serverConfigTest.DataId,
		Group:   serverConfigTest.Group,
		Content: content})

	return success, err
}
func getNacosDataSource() *NacosDataSource {
	mh1 := &datasource.MockPropertyHandler{}
	mh1.On("Handle", tmock.Anything).Return(nil)
	mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
	nds := NewNacosDataSource(serverConfigTest, mh1)

	return nds
}

func TestNacosDataSource_Initialize(t *testing.T) {
	t.Run("NacosDataSource_Initialize_BuildNacosClient", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)

		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		nds := getNacosDataSource()
		err = nds.Initialize()

		assert.True(t, err == nil, "NacosDataSource initialize.")
	})
	t.Run("NacosDataSource_Initialize_listen", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)

		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		nds := getNacosDataSource()
		err = nds.Initialize()

		assert.True(t, err == nil, "NacosDataSource initialize.")

		time.Sleep(2 * time.Second)
		nds.configClient.PublishConfig(vo.ConfigParam{
			DataId:  serverConfigTest.DataId,
			Group:   serverConfigTest.Group,
			Content: "123"})

		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")
		time.Sleep(2 * time.Second)
	})
}

func TestNacosDataSource_ReadSource(t *testing.T) {
	t.Run("NacosDataSource_ReadSource", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)

		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		nds := getNacosDataSource()
		err = nds.Initialize()

		assert.True(t, err == nil, "NacosDataSource initialize.")

		data, err := nds.ReadSource()

		assert.True(t, data != nil && err == nil, "NacosDataSource read source success.")
	})
	t.Run("NacosDataSource_ReadSource_Err", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)

		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		mh1 := &datasource.MockPropertyHandler{}
		mh1.On("Handle", tmock.Anything).Return(nil)
		mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
		nds := NewNacosDataSource(serverConfigErrTest, mh1)
		err = nds.Initialize()

		assert.True(t, err == nil, "NacosDataSource initialize.")

		data, err := nds.ReadSource()

		assert.True(t, data == nil && err != nil, "NacosDataSource read source failed.")
	})
}

func TestNacosDataSource_Close(t *testing.T) {
	published, err := prePushSystemRules(TestSystemRules)
	assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

	nds := getNacosDataSource()
	err = nds.Initialize()
	assert.True(t, err == nil, "NacosDataSource initialize.")

	err = nds.Close()
	assert.Nil(t, err)
}
