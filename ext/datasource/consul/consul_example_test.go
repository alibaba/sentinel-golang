package consul

import (
	"time"

	"github.com/hashicorp/consul/api"
)

func Example_consulDatasource_CustomizeClient() {
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
	})
	if err != nil {
		// todo something
	}
	ds, err := NewDatasource("property_key",
		// customize consul client
		WithConsulClient(client),
		// disable dynamic datasource watch
		WithDisableWatch(true),
		// preset property handlers
		WithPropertyHandlers(),
		// reset queryOptions, defaultQueryOptions as default
		WithQueryOptions(&api.QueryOptions{}),
	)

	if err != nil {
		// todo something
	}

	if err := ds.Initialize(); err != nil {
		// todo something
	}
}

func Example_consulDatasource_CustomizeConfig() {
	ds, err := NewDatasource("property_key",
		// customize consul config
		WithConsulConfig(&api.Config{
			Address: "127.0.0.1:8500",
		}),
		// disable dynamic datasource watch
		WithDisableWatch(true),
		// preset property handlers
		WithPropertyHandlers(),
		// reset queryOptions, defaultQueryOptions as default
		WithQueryOptions(&api.QueryOptions{
			WaitIndex: 0,
			// override default WaitTime(5min)
			WaitTime: time.Second * 90,
		}),
	)

	if err != nil {
		// todo something
	}

	if err := ds.Initialize(); err != nil {
		// todo something
	}
}
