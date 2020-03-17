package refreshable_file

import (
	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

var (
	logger = logging.GetDefaultLogger()
)

type RefreshableFileDataSource struct {
	datasource.Base
	sourceFilePath string
	isInitialized  util.AtomicBool
	closeChan      chan struct{}
	watcher        *fsnotify.Watcher
}

func NewFileDataSource(sourceFilePath string, handlers ...datasource.PropertyHandler) (*RefreshableFileDataSource, error) {
	var ds = &RefreshableFileDataSource{
		sourceFilePath: sourceFilePath,
		closeChan:      make(chan struct{}),
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds, ds.Initialize()
}

func (s *RefreshableFileDataSource) ReadSource() ([]byte, error) {
	f, err := os.Open(s.sourceFilePath)
	if err != nil {
		return nil, errors.Errorf("Fail to open the property file, err: %+v.", errors.WithStack(err))
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Errorf("Fail to read file, err: %+v.", errors.WithStack(err))
	}
	return src, nil
}

func (s *RefreshableFileDataSource) Initialize() error {
	if !s.isInitialized.CompareAndSet(false, true) {
		return nil
	}

	err := s.doReadAndUpdate()
	if err != nil {
		logger.Errorf("Fail to execute doReadAndUpdate, err: %+v", err)
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return errors.Errorf("Fail to new a watcher instance of fsnotify, err: %+v", err)
	}
	err = w.Add(s.sourceFilePath)
	if err != nil {
		return errors.Errorf("Fail add a watcher on file(%s), err: %+v", s.sourceFilePath, err)
	}
	s.watcher = w

	go util.RunWithRecover(func() {
		defer s.watcher.Close()
		for {
			select {
			case ev := <-s.watcher.Events:
				if ev.Op&fsnotify.Write == fsnotify.Write {
					err := s.doReadAndUpdate()
					if err != nil {
						logger.Errorf("Fail to execute doReadAndUpdate, err: %+v", err)
					}
				}

				if ev.Op&fsnotify.Remove == fsnotify.Remove || ev.Op&fsnotify.Rename == fsnotify.Rename {
					logger.Errorf("The file source(%s) was removed or renamed.", s.sourceFilePath)
					for _, h := range s.Handlers() {
						err := h.Handle(nil)
						if err != nil {
							logger.Errorf("RefreshableFileDataSource fail to publish property, err: %+v.", errors.WithStack(err))
						}
					}
				}
			case err := <-s.watcher.Errors:
				logger.Errorf("Watch err on file(%s), err: %+v", s.sourceFilePath, err)
			case <-s.closeChan:
				return
			}
		}
	}, logger)
	return nil
}

func (s *RefreshableFileDataSource) doReadAndUpdate() error {
	src, err := s.ReadSource()
	if err != nil {
		return errors.Errorf("Fail to read source, err: %+v", err)
	}
	for _, h := range s.Handlers() {
		err := h.Handle(src)
		if err != nil {
			return errors.Errorf("RefreshableFileDataSource fail to publish property, err: %+v.", err)
		}
	}
	return nil
}

func (s *RefreshableFileDataSource) Close() error {
	s.closeChan <- struct{}{}
	logger.Info("The RefreshableFileDataSource had been closed.")
	return nil
}
