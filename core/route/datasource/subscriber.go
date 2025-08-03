package datasource

import (
	"github.com/alibaba/sentinel-golang/core/route"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/opensergo/opensergo-go/pkg/api"
	"github.com/opensergo/opensergo-go/pkg/client"
	"github.com/opensergo/opensergo-go/pkg/common/logging"
	"github.com/opensergo/opensergo-go/pkg/configkind"
	"github.com/opensergo/opensergo-go/pkg/model"
	"github.com/pkg/errors"
)

type TrafficRouterSubscriber struct {
	OnUpdate func(subscribeKey model.SubscribeKey, data interface{}) (bool, error)
}

func (t TrafficRouterSubscriber) OnSubscribeDataUpdate(subscribeKey model.SubscribeKey, data interface{}) (bool, error) {
	ok, err := t.OnUpdate(subscribeKey, data)
	return ok, err
}

func SubscribeOpenSergoTrafficConfig(host, namespace, app string, port uint32) error {

	// Set OpenSergo console logger (optional)
	logging.NewConsoleLogger(logging.InfoLevel, logging.SeparateFormat, true)

	// Create an OpenSergoClient.
	openSergoClient, err := client.NewOpenSergoClient(host, port)
	if err != nil {
		return err
	}

	// Start the OpenSergoClient.
	err = openSergoClient.Start()
	if err != nil {
		return err
	}

	subscribeKey := model.NewSubscribeKey(namespace, app, configkind.ConfigKindTrafficRouterStrategy{})
	subscriber := TrafficRouterSubscriber{
		OnUpdate: func(subscribeKey model.SubscribeKey, data interface{}) (bool, error) {
			messages, ok := data.([]routev3.RouteConfiguration)
			if !ok {
				return false, errors.New("failed to convert data to RouteConfiguration")
			}
			TRList, VWList := resolveRouting(messages)
			route.SetTrafficRouterList(TRList)
			route.SetVirtualWorkloadList(VWList)
			return true, nil
		},
	}

	err = openSergoClient.SubscribeConfig(*subscribeKey, api.WithSubscriber(subscriber))
	if err != nil {
		return err
	}

	return err
}
