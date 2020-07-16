package nacos

import (
	"strings"
	"testing"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/nacos-group/nacos-sdk-go/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/common/http_agent"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/assert"
	tmock "github.com/stretchr/testify/mock"
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
	Port:        80,
	IpAddr:      "console.nacos.io",
}
var serverConfigErr = constant.ServerConfig{
	ContextPath: "/nacos",
	IpAddr:      "console.nacos.io1",
}
var clientConfigTest = constant.ClientConfig{
	TimeoutMs:      10000,
	ListenInterval: 20000,
	BeatInterval:   10000,
}

var configParam = ConfigParam{
	DataId: "system-rules",
	Group:  "sentinel-go",
}
var configParamErr = ConfigParam{
	Group: "sentinel-go",
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
		DataId:  configParam.DataId,
		Group:   configParam.Group,
		Content: content})

	return success, err
}
func getNacosDataSource(clientConfigTest constant.ClientConfig, serverConfig constant.ServerConfig, configParam ConfigParam) (*NacosDataSource, error) {
	mh1 := &datasource.MockPropertyHandler{}
	mh1.On("Handle", tmock.Anything).Return(nil)
	mh1.On("isPropertyConsistent", tmock.Anything).Return(false)
	nds, err := NewNacosDataSource(clientConfigTest, serverConfig, configParam, mh1)

	return nds, err
}

func TestNewNacosDataSource(t *testing.T) {
	t.Run("NewNacosDataSource", func(t *testing.T) {
		nds, err := getNacosDataSource(clientConfigTest, serverConfig, configParam)
		assert.True(t, nds != nil && err == nil, "New NacosDataSource success.")
	})
	t.Run("NewNacosDataSourceErr", func(t *testing.T) {
		mh1 := &datasource.MockPropertyHandler{}
		nds, err := NewNacosDataSource(clientConfigTest, serverConfigErr, configParam, mh1)
		assert.True(t, nds == nil && err != nil && strings.Contains(err.Error(), "The nacos serverConfig is incorrect."), "New NacosDataSource failed.")
	})
}

func TestNacosDataSource_Initialize(t *testing.T) {

	t.Run("NacosDataSource_Initialize_BuildNacosClient", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)
		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		nds, err := getNacosDataSource(clientConfigTest, serverConfig, configParam)
		assert.True(t, err == nil, "New NacosDataSource success.")
		err = nds.Initialize()
		assert.True(t, err == nil, "NacosDataSource initialize.")
	})

	t.Run("NacosDataSource_Initialize_BuildNacosClientErr", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)
		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		clientConfigTest.TimeoutMs = 0
		nds, err := getNacosDataSource(clientConfigTest, serverConfig, configParam)
		assert.True(t, err == nil, "New NacosDataSource success.")
		err = nds.Initialize()
		assert.True(t, err != nil && strings.Contains(err.Error(), "Nacosclient failed to build"), "NacosDataSource failed.")
	})
}

func TestNacosDataSource_ReadSource(t *testing.T) {
	t.Run("NacosDataSource_ReadSource", func(t *testing.T) {
		published, err := prePushSystemRules(TestSystemRules)
		assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

		nds, err := getNacosDataSource(clientConfigTest, serverConfig, configParam)
		assert.True(t, err == nil, "New NacosDataSource success.")
		err = nds.Initialize()
		assert.True(t, err == nil, "NacosDataSource initialize.")

		data, err := nds.ReadSource()
		assert.True(t, data != nil && err == nil, "NacosDataSource read source success.")
	})
}

func TestNacosDataSource_Close(t *testing.T) {
	published, err := prePushSystemRules(TestSystemRules)
	assert.True(t, err == nil && published, "Push systemRules configuration is successful.")

	nds, err := getNacosDataSource(clientConfigTest, serverConfig, configParam)
	assert.True(t, err == nil, "New NacosDataSource success.")
	err = nds.Initialize()
	assert.True(t, err == nil, "NacosDataSource initialize.")

	err = nds.Close()
	assert.Nil(t, err)
}
