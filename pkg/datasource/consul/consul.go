package consul

import (
	"context"
	"errors"
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/hashicorp/consul/api"
)

type consulDataSource struct {
	datasource.Base
	propertyKey   string
	kvQuerier     KVQuerier
	isInitialized util.AtomicBool
	cancel        context.CancelFunc
	queryOptions  api.QueryOptions
}

var (
	ErrNilConsulClient     = errors.New("nil consul client")
	ErrInvalidConsulConfig = errors.New("invalid consul config")
	ErrKeyDoesNotExist     = errors.New("key does not exist")
)

func NewDataSource(propertyKey string, opts ...Option) (datasource.DataSource, error) {
	var options = evaluateOptions(opts)
	// if user don't specify the consul client, sentinel should initialize from the configuration
	if options.consulClient == nil {
		if options.consulConfig == nil {
			return nil, ErrInvalidConsulConfig
		}
		client, err := api.NewClient(options.consulConfig)
		if err != nil {
			return nil, err
		}
		options.consulClient = client
	}

	// consul is still nil.
	if options.consulClient == nil {
		return nil, ErrNilConsulClient
	}
	return newConsulDataSource(propertyKey, options), nil
}

func newConsulDataSource(propertyKey string, options *options) *consulDataSource {
	ctx, cancel := context.WithCancel(options.queryOptions.Context())
	ds := &consulDataSource{
		propertyKey:  propertyKey,
		kvQuerier:    options.consulClient.KV(),
		cancel:       cancel,
		queryOptions: *options.queryOptions.WithContext(ctx),
	}

	for _, h := range options.propertyHandlers {
		ds.AddPropertyHandler(h)
	}
	return ds
}

func (c *consulDataSource) ReadSource() ([]byte, error) {
	pair, meta, err := c.kvQuerier.Get(c.propertyKey, &c.queryOptions)

	if err != nil {
		return nil, err
	}

	c.queryOptions.WaitIndex = meta.LastIndex
	if pair == nil {
		return nil, ErrKeyDoesNotExist
	}
	return pair.Value, nil
}

// Initialize implement datasource.DataSource interface
func (c *consulDataSource) Initialize() error {
	if !c.isInitialized.CompareAndSet(false, true) {
		return errors.New("consul datasource had been initialized")
	}
	if err := c.doReadAndUpdate(); err != nil {
		// Failed to read default should't block initialization
		logging.Error(err, "Failed to read initial data for key in consulDataSource.Initialize()", "propertyKey", c.propertyKey)
	}

	go util.RunWithRecover(c.watch)

	return nil
}

func (c *consulDataSource) watch() {
	logging.Info("[Consul] Consul data source is watching property", "propertyKey", c.propertyKey)
	for {
		if err := c.doReadAndUpdate(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, ErrKeyDoesNotExist) {
				continue
			}

			if api.IsRetryableError(err) {
				logging.Warn("[Consul] Update failed with retryable error in consulDataSource.watch()", "err", err)
				time.Sleep(time.Second)
				continue
			}
			logging.FrequentErrorOnce.Do(func() {
				logging.Error(err, "Failed to update data in consulDataSource.watch()", "propertyKey", c.propertyKey)
			})
		}
	}
}

func (c *consulDataSource) doUpdate(src []byte) (err error) {
	if len(src) == 0 {
		return c.Handle(nil)
	}
	return c.Handle(src)
}

func (c *consulDataSource) doReadAndUpdate() (err error) {
	src, err := c.ReadSource()
	if err != nil {
		return err
	}
	return c.doUpdate(src)
}

func (c *consulDataSource) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	logging.Info("[Consul] Consul data source has been closed", "propertyKey", c.propertyKey)
	return nil
}
