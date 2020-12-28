// Copyright 1999-2020 Alibaba Group Holding Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package file

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/alibaba/sentinel-golang/ext/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
)

type RefreshableFileDataSource struct {
	datasource.Base
	sourceFilePath string
	isInitialized  util.AtomicBool
	closeChan      chan struct{}
	watcher        *fsnotify.Watcher
	closed         util.AtomicBool
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
		logging.Error(err, "Fail to execute RefreshableFileDataSource.doReadAndUpdate")
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
				if ev.Op&fsnotify.Rename == fsnotify.Rename {
					logging.Warn("[RefreshableFileDataSource] The file source was renamed.", "sourceFilePath", s.sourceFilePath)
					updateErr := s.Handle(nil)
					if updateErr != nil {
						logging.Error(updateErr, "Fail to update nil property")
					}

					// try to watch sourceFile
					_ = s.watcher.Remove(s.sourceFilePath)
					retryCount := 0
					for {
						if retryCount > 5 {
							logging.Error(errors.New("retry failed"), "Fail to retry watch", "sourceFilePath", s.sourceFilePath)
							s.Close()
							return
						}
						e := s.watcher.Add(s.sourceFilePath)
						if e == nil {
							break
						}
						retryCount++
						logging.Error(e, "Failed to add to watcher", "sourceFilePath", s.sourceFilePath)
						util.Sleep(time.Second)
					}
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					logging.Warn("[RefreshableFileDataSource] The file source was removed.", "sourceFilePath", s.sourceFilePath)
					updateErr := s.Handle(nil)
					if updateErr != nil {
						logging.Error(updateErr, "Fail to update nil property")
					}
					s.Close()
					return
				}

				err := s.doReadAndUpdate()
				if err != nil {
					logging.Error(err, "Fail to execute RefreshableFileDataSource.doReadAndUpdate")
				}
			case err := <-s.watcher.Errors:
				logging.Error(err, "Watch err on file", "sourceFilePath", s.sourceFilePath)
			case <-s.closeChan:
				return
			}
		}
	})
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
	if !s.closed.CompareAndSet(false, true) {
		return nil
	}
	s.closeChan <- struct{}{}
	logging.Info("[File] The RefreshableFileDataSource for file had been closed.", "sourceFilePath", s.sourceFilePath)
	return nil
}
