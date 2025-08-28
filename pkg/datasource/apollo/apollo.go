package apollo

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/apolloconfig/agollo/v4"
	"github.com/apolloconfig/agollo/v4/component/log"
	"github.com/apolloconfig/agollo/v4/env/config"
	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/pkg/errors"
)

var (
	ErrEmptyKey   = errors.New("property key is empty")
	ErrMissConfig = errors.New("miss config")
	ErrConfigNil  = errors.New("cant find a config by the special namespace")
)

const defaultNamespace = "application"

type Option func(o *options)

type options struct {
	handlers  []datasource.PropertyHandler
	logger    log.LoggerInterface
	namespace string
}

// WithPropertyHandlers set property handlers
func WithPropertyHandlers(handlers ...datasource.PropertyHandler) Option {
	return func(o *options) {
		o.handlers = handlers
	}
}

// WithLogger set apollo logger
func WithLogger(logger log.LoggerInterface) Option {
	return func(o *options) {
		o.logger = logger
	}
}

// WithNamespace set apollo namespace to supports fetching rules from specified namespace.
func WithNamespace(namespace string) Option {
	return func(o *options) {
		o.namespace = namespace
	}
}

// apolloDatasource implements datasource.Datasource
type apolloDatasource struct {
	datasource.Base
	client      apolloClient
	propertyKey string
	namespace   string
}

// apolloClient a simple apollo client to support sentinel datasource
type apolloClient interface {
	GetConfig(namespace string) *storage.Config
	AddChangeListener(listener storage.ChangeListener)
}

// NewDatasource create apollo datasource
func NewDatasource(conf *config.AppConfig, propertyKey string, opts ...Option) (datasource.DataSource, error) {
	if conf == nil {
		return nil, ErrMissConfig
	}
	if propertyKey == "" {
		return nil, ErrEmptyKey
	}
	option := &options{
		logger: &log.DefaultLogger{},
	}
	for _, opt := range opts {
		opt(option)
	}
	if option.namespace == "" {
		option.namespace = defaultNamespace
	}
	agollo.SetLogger(option.logger)
	apolloClient, err := agollo.StartWithConfig(func() (*config.AppConfig, error) {
		return conf, nil
	})
	if err != nil {
		return nil, err
	}
	ds := &apolloDatasource{
		client:      apolloClient,
		propertyKey: propertyKey,
		namespace:   option.namespace,
	}
	for _, handler := range option.handlers {
		ds.AddPropertyHandler(handler)
	}
	return ds, nil
}

func (a *apolloDatasource) ReadSource() ([]byte, error) {
	cfg := a.client.GetConfig(a.namespace)
	if cfg == nil {
		return nil, ErrConfigNil
	}
	value := cfg.GetValue(a.propertyKey)
	return []byte(value), nil
}

func (a *apolloDatasource) Initialize() error {
	source, err := a.ReadSource()
	if err != nil {
		return err
	}
	a.handle(source)
	listener := &customChangeListener{
		ds: a,
	}
	a.client.AddChangeListener(listener)
	return nil
}

func (a *apolloDatasource) Close() error {
	return nil
}

func (a *apolloDatasource) handle(source []byte) {
	err := a.Handle(source)
	if err != nil {
		log.Errorf("update config err: %s", err.Error())
	}
}
