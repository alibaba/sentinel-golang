package apollo

import (
	"testing"

	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/stretchr/testify/assert"
)

// must run an apollo server before test
// this would be failed for CICD test. so it's disabled for now
//func TestConfig(t *testing.T) {
//
//	var resultConfig string
//
//	propertyHandler := datasource2.NewDefaultPropertyHandler(
//		func(src []byte) (interface{}, error) {
//			return string(src), nil
//		},
//		func(data interface{}) error {
//			s := data.(string)
//			fmt.Println(s)
//			resultConfig = s
//			return nil
//		},
//	)
//	c := &config.AppConfig{
//		AppID:          "SampleApp",
//		Cluster:        "DEV",
//		IP:             "http://localhost:8080",
//		NamespaceName:  "application",
//		IsBackupConfig: true,
//		Secret:         "1dc9532d02cd47f0bb26ee39d614ef8d",
//	}
//	datasource, err := NewDatasource(
//		c, "timeout",
//		WithLogger(&log.DefaultLogger{}),
//		WithPropertyHandlers(propertyHandler),
//	)
//	assert.Nil(t, err)
//	err = datasource.Initialize()
//	assert.Nil(t, err)
//	assert.Equal(t, "123", resultConfig)
//	select {}
//}

func TestEmptyKey(t *testing.T) {
	c := &config.AppConfig{
		AppID:          "SampleApp",
		Cluster:        "DEV",
		IP:             "http://localhost:8080",
		NamespaceName:  "application",
		IsBackupConfig: true,
		Secret:         "1dc9532d02cd47f0bb26ee39d614ef8d",
	}
	_, err := NewDatasource(c, "")
	assert.Equal(t, ErrEmptyKey, err)
}

func TestMissConfig(t *testing.T) {
	_, err := NewDatasource(nil, "test")
	assert.Equal(t, ErrMissConfig, err)
}
