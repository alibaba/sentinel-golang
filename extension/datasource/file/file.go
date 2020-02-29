package file

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/extension/datasource"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/fsnotify/fsnotify"
)

type fileDataSource struct {
	datasource.Base
	path string
	dataFormat string

	enableWatcher bool
}

func New(path string) base.DataSource {
	ds :=  &fileDataSource{
		Base: datasource.Base{DataFormat:filepath.Ext(path)},
		path: path,
		enableWatcher: true,
	}
	
	fmt.Printf("ds = %+v\n", ds)

	return ds
}

// ReadConfig implements base.DataSource interface
func (ds fileDataSource) ReadConfig() error {
	if err := ds.loadConfig(); err != nil {
		return err
	}

	if ds.enableWatcher {
		go ds.watch()
	}

	return nil
}

// Close implements base.DataSource interface
func (ds fileDataSource) Close() error {
	// note(gorexlv): don't close file here, see loadConfig
	return nil
}

func (ds fileDataSource) loadConfig() error {
	file, err := os.Open(ds.path)
	if err != nil {
		return err
	}

	defer func() {
		// close immediately, will not old this file for long term
		file.Close()
	}()

	src, err := ioutil.ReadAll(file)
	if err != nil {
		return nil
	}
	return ds.ApplyConfig(src)
}

func (ds fileDataSource) buildDecoder(reader io.Reader) base.PropertyDecoder {
	switch ds.dataFormat {
	case ".json":
		return json.NewDecoder(reader)
	case ".yaml":
		return json.NewDecoder(reader)
	case ".ini", ".toml", ".hcl":
		panic("unsupported data format by now")
	}
	panic("invalid data format")
}

func (ds fileDataSource) watch() {
	for {
		watch, err := fsnotify.NewWatcher()
		if err != nil {
			logging.GetDefaultLogger().Error("watch file", ds.path, "err", err, "event", "new")
			time.Sleep(time.Second)
			continue
		}
		err = watch.Add(ds.path)
		if err != nil {
			logging.GetDefaultLogger().Error("watch file", ds.path, "err", err, "event", "watch")
			time.Sleep(time.Second)
			continue
		}
		for {
			select {
			case ev := <-watch.Events:
				if ev.Op&fsnotify.Write == fsnotify.Write {
					if err := ds.loadConfig(); err != nil {
						logging.GetDefaultLogger().Error("watch file", ds.path, "err", err, "event", "write")
					}
				}
				if ev.Op&fsnotify.Remove == fsnotify.Remove {
					if err := ds.DeleteConfig(); err != nil {
						logging.GetDefaultLogger().Error("watch file", ds.path, "err", err, "event", "remove")
					}
				}
			case err := <-watch.Errors:
				logging.GetDefaultLogger().Error("watch file", ds.path, "err", err, "event", "watch")
				continue
			}
		}
	}
}

