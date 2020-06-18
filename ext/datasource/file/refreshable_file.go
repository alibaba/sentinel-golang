package file

import (
	"io/ioutil"
	"os"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
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

func NewFileDataSource(sourceFilePath string, handlers ...datasource.PropertyHandler) *RefreshableFileDataSource {
	var ds = &RefreshableFileDataSource{
		sourceFilePath: sourceFilePath,
		closeChan:      make(chan struct{}),
	}
	for _, h := range handlers {
		ds.AddPropertyHandler(h)
	}
	return ds
}

func (s *RefreshableFileDataSource) ReadSource() ([]byte, error) {
	f, err := os.Open(s.sourceFilePath)
	if err != nil {
		return nil, errors.Errorf("RefreshableFileDataSource fail to open the property file, err: %+v.", err)
	}
	defer f.Close()

	src, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, errors.Errorf("RefreshableFileDataSource fail to read file, err: %+v.", err)
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
		return errors.Errorf("Fail add a watcher on file[%s], err: %+v", s.sourceFilePath, err)
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
					logger.Warnf("The file source [%s] was removed or renamed.", s.sourceFilePath)
					updateErr := s.Handle(nil)
					if updateErr != nil {
						logger.Errorf("Fail to update nil property, err: %+v", updateErr)
					}
				}
			case err := <-s.watcher.Errors:
				logger.Errorf("Watch err on file[%s], err: %+v", s.sourceFilePath, err)
			case <-s.closeChan:
				return
			}
		}
	}, logger)
	return nil
}

func (s *RefreshableFileDataSource) doReadAndUpdate() (err error) {
	src, err := s.ReadSource()
	if err != nil {
		err = errors.Errorf("Fail to read source, err: %+v", err)
		return err
	}
	return s.Handle(src)
}

func (s *RefreshableFileDataSource) Close() error {
	s.closeChan <- struct{}{}
	logger.Infof("The RefreshableFileDataSource for [%s] had been closed.", s.sourceFilePath)
	return nil
}
