package nacos_v2

import (
	"errors"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"sync"
)

type NacosDataSource struct {
	client        config_client.IConfigClient
	isInitialized util.AtomicBool
	namespaceId   string
	dataId        sync.Map
}

func NewNacosDataSource(serverHost, namespaceId string, conf *Config) (*NacosDataSource, error) {
	if serverHost == "" || namespaceId == "" {
		return nil, errors.New("server host or namespace is empty")
	}

	// init nacos config client
	clientConfig := &constant.ClientConfig{
		NamespaceId: namespaceId,
		Endpoint:    serverHost,
		TimeoutMs:   DefaultTimeoutMs,
	}
	if conf != nil {
		if conf.TimeoutMs != 0 {
			clientConfig.TimeoutMs = conf.TimeoutMs
		}
	}
	client, err := clients.NewConfigClient(vo.NacosClientParam{
		ClientConfig: clientConfig,
	})
	if err != nil {
		return nil, err
	}

	return &NacosDataSource{
		client:      client,
		namespaceId: namespaceId,
		dataId:      sync.Map{},
	}, nil
}

func (n *NacosDataSource) RegisterRuleDataSource(group, dataId string, handler func(string, string, string, string), preprocess func(string, string, string, string) (string, error)) error {
	if v, ok := n.dataId.LoadOrStore(dataId, false); ok && v == true {
		return nil
	}
	if !n.dataId.CompareAndSwap(dataId, false, true) {
		return nil
	}

	var onChangeHandler func(namespace, group, dataId, data string)
	if handler == nil {
		handler = defaultOnRuleChangeHandler
	}

	if preprocess == nil {
		onChangeHandler = handler
	} else {
		onChangeHandler = func(namespace, group, dataId, data string) {
			newData, err := preprocess(namespace, group, dataId, data)
			if err != nil {
				logging.Error(err, "Failed to preprocess data", "dataId", dataId)
				return
			}
			handler(namespace, group, dataId, newData)
		}
	}

	config := vo.ConfigParam{
		Group:    group,
		DataId:   dataId,
		OnChange: onChangeHandler,
	}
	go func() {
		data, err := n.client.GetConfig(config)
		if err != nil && err.Error() != "config data not exist" {
			logging.Error(err, "Failed to getConfig from ACM", "dataId", dataId)
		} else if len(data) > 0 {
			onChangeHandler(n.namespaceId, group, dataId, data)
		}
	}()
	return n.client.ListenConfig(config)
}
