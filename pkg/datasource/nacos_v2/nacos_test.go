package nacos_v2

import (
	"fmt"
	"testing"
)

func TestNacosDataSource(t *testing.T) {
	nacosDataSourc, err := NewNacosDataSource("-", "-", nil)
	if err != nil {
		t.Error(err)
	}
	err = nacosDataSourc.RegisterRuleDataSource("ahas-sentinel-cn-hangzhou-online", "flow-rule-1784327288677274-zyx-governance-demo-gin-server-b", nil, preprocessHandler)
	if err != nil {
		t.Error(err)
	}

	select {}
}

func preprocessHandler(namespace, group, dataId, data string) (string, error) {
	fmt.Printf("preprocessHandler: %s\n", data)
	return data, nil
}
