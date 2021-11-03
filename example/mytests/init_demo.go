package main

import (
	"encoding/json"
	"fmt"
	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/config"
)

func main() {

	parser := func(configBytes []byte) (*config.Entity, error) {
		conf := &config.Entity{}
		err := json.Unmarshal(configBytes, conf)
		return conf, err
	}
	conf := "{\"Version\":\"v1\",\"Sentinel\":{\"App\":{\"Name\":\"roshi-app\",\"Type\":0}}}"
	err := api.InitWithParser([]byte(conf), parser)

	fmt.Println(err)
}
