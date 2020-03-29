package consul

import (
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/hashicorp/consul/api"
)

type (
	options struct {
		consulConfig     *api.Config
		consulClient     *api.Client
		propertyHandlers []datasource.PropertyHandler
		queryOptions     *api.QueryOptions
		disableWatch     bool
	}

	Option func(*options)

	KVQuerier interface {
		Get(key string, q *api.QueryOptions) (*api.KVPair, *api.QueryMeta, error)
		List(prefix string, q *api.QueryOptions) (api.KVPairs, *api.QueryMeta, error)
	}
)

// WithQueryOptions set options for consulClient.Get method
func WithQueryOptions(queryOptions *api.QueryOptions) Option {
	return func(opts *options) {
		opts.queryOptions = queryOptions
	}
}

// WithConsulConfig inject consul client config
func WithConsulConfig(config *api.Config) Option {
	return func(opts *options) {
		opts.consulConfig = config
	}
}

// WithConsulClient inject consul client instance
func WithConsulClient(client *api.Client) Option {
	return func(opts *options) {
		opts.consulClient = client
	}
}

// WithPropertyHandlers inject property handlers
func WithPropertyHandlers(handlers ...datasource.PropertyHandler) Option {
	return func(opts *options) {
		if opts.propertyHandlers == nil {
			opts.propertyHandlers = make([]datasource.PropertyHandler, 0)
		}
		opts.propertyHandlers = append(opts.propertyHandlers, handlers...)
	}
}

// WithDisableWatch disable watch
func WithDisableWatch(disableWatch bool) Option {
	return func(opts *options) {
		opts.disableWatch = disableWatch
	}
}

func evaluateOptions(opts []Option) *options {
	var optCopy = &options{
		propertyHandlers: make([]datasource.PropertyHandler, 0),
		// default query options
		queryOptions: defaultQueryOptions(),
		// enable watch be default
		disableWatch: false,
	}
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

func defaultQueryOptions() *api.QueryOptions {
	return &api.QueryOptions{
		Datacenter:        "",
		AllowStale:        false,
		RequireConsistent: false,
		UseCache:          false,
		MaxAge:            0,
		StaleIfError:      0,
		WaitIndex:         0,
		WaitHash:          "",
		WaitTime:          time.Minute * 5, // block request
		Token:             "",
		Near:              "",
		NodeMeta:          nil,
		RelayFactor:       0,
		Connect:           false,
		Filter:            "",
	}
}
