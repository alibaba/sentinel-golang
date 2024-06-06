package xds

import (
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/bootstrap"
	"github.com/alibaba/sentinel-golang/pkg/datasource/xds/utils"
	"os"
)

var XdsAgent *Agent

func init() {
	var err error
	node, err := bootstrap.InitNode()
	if err != nil {
		logging.Error(err, "init xds agent InitNode failed")
		XdsAgent = &Agent{}
		return
	}

	xdsServerAddr := os.Getenv(utils.EnvIstioAddress)
	if xdsServerAddr == "" {
		logging.Warn("init xds agent xdsServerAddr is empty")
		XdsAgent = &Agent{}
		return
	}

	XdsAgent, err = NewXdsAgent(xdsServerAddr, node)
	if err != nil {
		logging.Error(err, "init xds agent NewXdsAgent failed")
		XdsAgent = &Agent{}
		return
	}
}
