package apollo

import (
	"errors"
	"fmt"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/shima-park/agollo"
)

type apolloDataSource struct {
	datasource.Base
	namespace     string
	client        agollo.Agollo
	isInitialized util.AtomicBool
	stop          chan bool
}

func NewDatasource(client agollo.Agollo, namespace string, handlers ...datasource.PropertyHandler) (datasource.DataSource, error) {

	if namespace == "" {
		return nil, errors.New("The namespace is empty.")
	}

	if client == nil {
		return nil, errors.New("The agollo client is nil.")
	}

	ds := &apolloDataSource{
		client:    client,
		namespace: namespace,
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds, nil
}

func (s *apolloDataSource) Initialize() error {
	if !s.isInitialized.CompareAndSet(false, true) {
		return errors.New("Apollo datasource had been initialized")
	}
	if err := s.doReadAndUpdate(); err != nil {
		return fmt.Errorf("Failed to read initial data: %v", err)
	}

	go util.RunWithRecover(s.watch)

	return nil
}

func (s *apolloDataSource) ReadSource() ([]byte, error) {
	raw := s.client.GetNameSpace(s.namespace)
	v, exist := raw["content"]
	if !exist {
		return nil, errors.New("namespace does not exist/empty")
	}

	str, ok := v.(string)
	if !ok {
		return nil, errors.New("namespace val assert failed")
	}

	return []byte(str), nil
}

func (s *apolloDataSource) Close() error {
	s.client.Stop()
	close(s.stop)
	return nil
}

func (s *apolloDataSource) watch() {

	errChan := s.client.Start()
	watchNSCh := s.client.WatchNamespace(s.namespace, s.stop)
	for {
		select {
		case <-s.stop:
			return
		case err := <-errChan:
			logging.Error(err.Err, "ApolloDataSource: watch")
		case <-watchNSCh:
			s.doReadAndUpdate()
		}
	}
}

func (s *apolloDataSource) doReadAndUpdate() error {
	src, err := s.ReadSource()
	if err != nil {
		return err
	}

	return s.Handle(src)
}
