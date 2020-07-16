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
	ErrKeyIsNotExisted     = errors.New("key is not existed now")

	logger = logging.GetDefaultLogger()
)

func NewDatasource(propertyKey string, opts ...Option) (datasource.DataSource, error) {
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
		return nil, ErrKeyIsNotExisted
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
		logger.Errorf("[consul] fail to initialize key: %s, err: %s", c.propertyKey, err.Error())
	}

	go util.RunWithRecover(c.watch, logger)

	return nil
}

func (c *consulDataSource) watch() {
	logger.Infof("consul datasource is watching property: %s", c.propertyKey)
	for {
		if err := c.doReadAndUpdate(); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, ErrKeyIsNotExisted) {
				continue
			}

			if api.IsRetryableError(err) {
				logger.Warnf("[consul] doUpdate failed with retryable error: %s", err.Error())
				time.Sleep(time.Second)
				continue
			}

			logger.Errorf("[consul] doUpdate failed: %s", err.Error())
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
	logger.Info("[consul] close consul datasource")
	return nil
}
