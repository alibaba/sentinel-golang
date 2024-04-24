package xds

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/xds/bootstrap"
)

var XdsAgent *Agent

func init() {
	// TODO: 通过open api获取xds地址
	var err error
	node, err := bootstrap.InitNode()
	if err != nil {
		fmt.Printf("init xds agent InitNode failed, err: %v\n", err)
		XdsAgent = &Agent{}
		return
	}

	XdsAgent, err = NewXdsAgent("47.97.99.1:15010", node)
	if err != nil {
		fmt.Printf("init xds agent NewXdsAgent failed, err: %v\n", err)
		XdsAgent = &Agent{}
		return
	}
}
