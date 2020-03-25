package etcdv3

import (
	"context"
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	"go.uber.org/multierr"
	"time"
)

var logger = logging.GetDefaultLogger()

type Etcdv3DataSource struct {
	datasource.Base
	propertyKey         string
	lastUpdatedRevision int64
	client              *clientv3.Client
	createdBy           *DatasourceGenerator
	closeChan           chan struct{}
}

func (s *Etcdv3DataSource) ReadSource() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := s.client.Get(ctx, s.propertyKey)
	if err != nil {
		return nil, err
	}
	if resp.Count == 0 {
		return nil, errors.Errorf("The key[%s] is not existed in etcd.", s.propertyKey)
	}
	s.lastUpdatedRevision = resp.Header.GetRevision()
	return resp.Kvs[0].Value, nil
}

func (s *Etcdv3DataSource) Initialize() error {
	err := s.doReadAndUpdate()
	if err != nil {
		return err
	}
	go util.RunWithRecover(s.watch, logger)
	return err
}

func (s *Etcdv3DataSource) doReadAndUpdate() error {
	src, err := s.ReadSource()
	if err != nil {
		err = errors.Errorf("Fail to read source, err: %+v", err)
		return err
	}
	for _, h := range s.Handlers() {
		e := h.Handle(src)
		if e != nil {
			err = multierr.Append(err, e)
		}
	}
	return err
}

func (s *Etcdv3DataSource) processWatchResponse(resp *clientv3.WatchResponse) {
	if resp.CompactRevision > s.lastUpdatedRevision {
		s.lastUpdatedRevision = resp.CompactRevision
	}
	if resp.Header.GetRevision() > s.lastUpdatedRevision {
		s.lastUpdatedRevision = resp.Header.GetRevision()
	}

	if err := resp.Err(); err != nil {
		logger.Errorf("Watch on etcd endpoints(%+v) occur error, err: %+v", s.client.Endpoints(), err)
		return
	}

	for _, ev := range resp.Events {
		if ev.Type == mvccpb.PUT {
			err := s.doReadAndUpdate()
			if err != nil {
				logger.Errorf("Fail to execute doReadAndUpdate for PUT event, err: %+v", err)
			}
		}
		if ev.Type == mvccpb.DELETE {
			var updateErr error
			for _, h := range s.Handlers() {
				e := h.Handle(nil)
				if e != nil {
					updateErr = multierr.Append(updateErr, e)
				}
			}
			if updateErr != nil {
				logger.Errorf("Fail to execute doReadAndUpdate for DELETE event, err: %+v", updateErr)
			}
		}
	}
}

func (s *Etcdv3DataSource) watch() {
	// Add watch for propertyKey from lastUpdatedRevision updated after Initializing
	rch := s.client.Watch(context.Background(), s.propertyKey, clientv3.WithRev(int64(s.lastUpdatedRevision)))
	for {
		select {
		case wresp := <-rch:
			s.processWatchResponse(&wresp)
		case <-s.closeChan:
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
		if s.lastUpdatedRevision > 0 {
			rch = s.client.Watch(context.Background(), s.propertyKey, clientv3.WithRev(s.lastUpdatedRevision))
		} else {
			rch = s.client.Watch(context.Background(), s.propertyKey)
		}
	}
}

func (s *Etcdv3DataSource) Close() error {
	// stop to watch property key.
	s.closeChan <- struct{}{}
	if s.createdBy.closeable() {
		return s.createdBy.close()
	}
	return nil
}
