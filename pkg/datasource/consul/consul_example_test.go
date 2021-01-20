package consul

import (
	"fmt"
	"os"
	"time"

	"github.com/hashicorp/consul/api"
)

func Example_consulDataSource_CustomizeClient() {
	client, err := api.NewClient(&api.Config{
		Address: "127.0.0.1:8500",
	})
	if err != nil {
		fmt.Println("Failed to instance consul client")
		os.Exit(1)
	}
	ds, err := NewDataSource("property_key",
		// customize consul client
		WithConsulClient(client),
		// preset property handlers
		WithPropertyHandlers(),
		// reset queryOptions, defaultQueryOptions as default
		WithQueryOptions(&api.QueryOptions{}),
	)

	if err != nil {
		fmt.Println("Failed to instance consul datasource")
		os.Exit(1)
	}

	if err := ds.Initialize(); err != nil {
		fmt.Println("Failed to initialize consul datasource")
		os.Exit(1)
	}
}

func Example_consulDataSource_CustomizeConfig() {
	ds, err := NewDataSource("property_key",
		// customize consul config
		WithConsulConfig(&api.Config{
			Address: "127.0.0.1:8500",
		}),
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
		fmt.Println("Failed to instance consul datasource")
		os.Exit(1)
	}

	if err := ds.Initialize(); err != nil {
		fmt.Println("Failed to initialize consul datasource")
		os.Exit(1)
	}
}
