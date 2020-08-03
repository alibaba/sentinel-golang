package nacos

import (
	"testing"

	"github.com/nacos-group/nacos-sdk-go/clients/config_client"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/stretchr/testify/assert"

	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"github.com/stretchr/testify/mock"
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

var (
	Group  = "sentinel-go"
	DataId = "system-rules"
)

type nacosClientMock struct {
	mock.Mock
}

func (n *nacosClientMock) GetConfig(param vo.ConfigParam) (string, error) {
	ret := n.Called(param)
	return ret.String(0), ret.Error(1)
}

func (n *nacosClientMock) PublishConfig(param vo.ConfigParam) (bool, error) {
	ret := n.Called(param)
	return ret.Bool(0), ret.Error(1)
}

func (n *nacosClientMock) DeleteConfig(param vo.ConfigParam) (bool, error) {
	ret := n.Called(param)
	return ret.Bool(0), ret.Error(1)
}

func (n *nacosClientMock) ListenConfig(params vo.ConfigParam) (err error) {
	ret := n.Called(params)
	return ret.Error(0)
}

func (n *nacosClientMock) CancelListenConfig(params vo.ConfigParam) (err error) {
	ret := n.Called(params)
	return ret.Error(0)
}

func (n *nacosClientMock) SearchConfig(param vo.SearchConfigParm) (*model.ConfigPage, error) {
	ret := n.Called(param)
	return ret.Get(0).(*model.ConfigPage), ret.Error(1)
}

func getNacosDataSource(client config_client.IConfigClient) (*NacosDataSource, error) {
	mh1 := &datasource.MockPropertyHandler{}
	mh1.On("Handle", mock.Anything).Return(nil)
	mh1.On("isPropertyConsistent", mock.Anything).Return(false)
	nds, err := NewNacosDataSource(client, Group, DataId, mh1)

	return nds, err
}

func TestNacosDataSource(t *testing.T) {

	t.Run("NewNacosDataSource", func(t *testing.T) {
		client, err := createConfigClientTest()
		assert.Nil(t, err)
		nds, err := getNacosDataSource(client)
		assert.True(t, nds != nil && err == nil)
	})

	t.Run("NacosDataSource_Initialize", func(t *testing.T) {
		mh1 := &datasource.MockPropertyHandler{}
		mh1.On("Handle", mock.Anything).Return(nil)
		mh1.On("isPropertyConsistent", mock.Anything).Return(false)
		nacosClientMock := new(nacosClientMock)
		nacosClientMock.On("GetConfig", mock.Anything).Return(TestSystemRules, nil)
		nacosClientMock.On("ListenConfig", mock.Anything).Return(nil)
		nds, err := getNacosDataSource(nacosClientMock)
		assert.True(t, nds != nil && err == nil)
		err = nds.Initialize()
		assert.True(t, err == nil)
	})
}
