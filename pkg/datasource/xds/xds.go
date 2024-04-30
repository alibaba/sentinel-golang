package xds

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/bootstrap"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/utils"
	"os"
)

var XdsAgent *Agent

func init() {
	var err error
	node, err := bootstrap.InitNode()
	if err != nil {
		fmt.Printf("init xds agent InitNode failed, err: %v\n", err)
		XdsAgent = &Agent{}
		return
	}

	xdsServerAddr := os.Getenv(utils.EnvIstioAddress)
	if xdsServerAddr == "" {
		fmt.Printf("init xds agent xdsServerAddr is empty\n")
		XdsAgent = &Agent{}
		return
	}

	XdsAgent, err = NewXdsAgent(xdsServerAddr, node)
	if err != nil {
		fmt.Printf("init xds agent NewXdsAgent failed, err: %v\n", err)
		XdsAgent = &Agent{}
		return
	}
}
