package etcdv3

import (
	"context"
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
)

type Etcdv3DataSource struct {
	datasource.Base
	propertyKey         string
	lastUpdatedRevision int64
	client              *clientv3.Client
	// cancel is the func, call cancel will stop watching on the propertyKey
	cancel context.CancelFunc
	// closed indicate whether continuing to watch on the propertyKey
	closed util.AtomicBool
}

// NewDataSource new a Etcdv3DataSource instance.
// client is the etcdv3 client, it must be useful and should be release by User.
func NewDataSource(client *clientv3.Client, key string, handlers ...datasource.PropertyHandler) (*Etcdv3DataSource, error) {
	if client == nil {
		return nil, errors.New("The etcdv3 client is nil.")
	}
	ds := &Etcdv3DataSource{
		client:      client,
		propertyKey: key,
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds, nil
}

func (s *Etcdv3DataSource) Initialize() error {
	err := s.doReadAndUpdate()
	if err != nil {
		logging.Error(err, "Fail to update data for key when execute Etcdv3DataSource.Initialize()", "propertyKey", s.propertyKey)
	}
	go util.RunWithRecover(s.watch)
	return nil
}

func (s *Etcdv3DataSource) ReadSource() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	resp, err := s.client.Get(ctx, s.propertyKey)
	if err != nil {
		return nil, errors.Errorf("Fail to get value for property key[%s]", s.propertyKey)
	}
	if resp.Count == 0 {
		return nil, errors.Errorf("The key[%s] is not existed in etcd server.", s.propertyKey)
	}
	s.lastUpdatedRevision = resp.Header.GetRevision()
	logging.Info("[Etcdv3] Get the newest data for key", "propertyKey", s.propertyKey,
		"revision", resp.Header.GetRevision(), "value", resp.Kvs[0].Value)
	return resp.Kvs[0].Value, nil
}

func (s *Etcdv3DataSource) doReadAndUpdate() error {
	src, err := s.ReadSource()
	if err != nil {
		return err
	}
	return s.Handle(src)
}

func (s *Etcdv3DataSource) processWatchResponse(resp *clientv3.WatchResponse) {
	if resp.CompactRevision > s.lastUpdatedRevision {
		s.lastUpdatedRevision = resp.CompactRevision
	}
	if resp.Header.GetRevision() > s.lastUpdatedRevision {
		s.lastUpdatedRevision = resp.Header.GetRevision()
	}

	if err := resp.Err(); err != nil {
		logging.Error(err, "Watch on etcd endpoints occur error", "endpointd", s.client.Endpoints())
		return
	}

	for _, ev := range resp.Events {
		if ev.Type == mvccpb.PUT {
			err := s.doReadAndUpdate()
			if err != nil {
				logging.Error(err, "Fail to execute doReadAndUpdate for PUT event")
			}
		}
		if ev.Type == mvccpb.DELETE {
			updateErr := s.Handle(nil)
			if updateErr != nil {
				logging.Error(updateErr, "Fail to execute doReadAndUpdate for DELETE event")
			}
		}
	}
}

func (s *Etcdv3DataSource) watch() {
	// Add watch for propertyKey from lastUpdatedRevision updated after Initializing
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	rch := s.client.Watch(ctx, s.propertyKey, clientv3.WithCreatedNotify(), clientv3.WithRev(s.lastUpdatedRevision))
	for {
		for resp := range rch {
			s.processWatchResponse(&resp)
		}
		// Stop watching if datasource had been closed.
		if s.closed.Get() {
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
		ctx, cancel = context.WithCancel(context.Background())
		s.cancel = cancel
		if s.lastUpdatedRevision > 0 {
			rch = s.client.Watch(ctx, s.propertyKey, clientv3.WithRev(s.lastUpdatedRevision+1))
		} else {
			rch = s.client.Watch(ctx, s.propertyKey)
		}
	}
}

func (s *Etcdv3DataSource) Close() error {
	// stop to watch property key.
	s.closed.Set(true)
	s.cancel()

	return nil
}
