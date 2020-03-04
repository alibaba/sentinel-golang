package config

import (
	"fmt"
	"github.com/alibaba/sentinel-golang/core/constant"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)

var (
	globalCfg = NewDefaultConfig()
	initLogOnce sync.Once
)

func LoadFromYamlFile(filePath string) error {
	if filePath == constant.DefaultConfigFilename {
		if _, err := os.Stat(constant.DefaultConfigFilename); err != nil {
			//use default globalCfg.
			return nil
		}
	}
	_, err := os.Stat(filePath)
	if err != nil && !os.IsExist(err){
		return err
	}
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(content, globalCfg)
	if err != nil {
		return err
	}
	logging.GetDefaultLogger().Infof("Resolving Sentinel globalCfg from file: %s", filePath)
	return checkValid(&(globalCfg.Sentinel))
}

func OverrideFromSystemEnv() error{
	if appName := os.Getenv(constant.AppNameEnvKey); !util.IsBlank(appName) {
		globalCfg.Sentinel.App.Name = appName
	}

	if appTypeStr := os.Getenv(constant.AppTypeEnvKey); !util.IsBlank(appTypeStr) {
		appType, err := strconv.ParseInt(appTypeStr, 10, 32)
		if err != nil {
			return err
		} else {
			globalCfg.Sentinel.App.Type = int32(appType)
		}
	}

	if addPidStr := os.Getenv(constant.LogNamePidEnvKey); !util.IsBlank(addPidStr) {
		addPid, err := strconv.ParseBool(addPidStr)
		if err != nil {
			return err
		} else {
			globalCfg.Sentinel.Log.UsePid = addPid
		}
	}

	if logDir := os.Getenv(constant.LogDirEnvKey); !util.IsBlank(logDir) {
		if _, err := os.Stat(logDir); err != nil && !os.IsExist(err) {
			return err
		}
		globalCfg.Sentinel.Log.Dir = logDir
	}
	return checkValid(&(globalCfg.Sentinel))
}

func InitializeLogConfig(logDir string, usePid bool) (err error) {
	if logDir == "" {
		return errors.New("Invalid empty log path")
	}

	initLogOnce.Do(func() {
		if err = util.CreateDirIfNotExists(logDir); err != nil {
			return
		}
		err = reconfigureRecordLogger(logDir, usePid)
	})
	return err
}

func reconfigureRecordLogger(logBaseDir string, withPid bool) error {
	logDir := addSeparatorIfNeeded(logBaseDir)
	filePath := logDir + logging.RecordLogFileName
	if withPid {
		filePath = filePath + ".pid" + strconv.Itoa(os.Getpid())
	}

	defaultLogger := logging.GetDefaultLogger()
	if defaultLogger == nil {
		return errors.New("Unexpected state: defaultLogger == nil")
	}
	logFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}

	// Note: not thread-safe!
	logging.ResetDefaultLogger(log.New(logFile, "", log.LstdFlags), logging.DefaultNamespace)
	fmt.Println("INFO: log base directory is: " + logDir)

	return nil
}

func GetDefaultLogDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return addSeparatorIfNeeded(home) + logging.DefaultDirName
}

func addSeparatorIfNeeded(path string) string {
	s := string(os.PathSeparator)
	if !strings.HasSuffix(path, s) {
		return path + s
	}
	return path
}



func AppName() string {
	return globalCfg.Sentinel.App.Name
}

func AppType() int32 {
	return globalCfg.Sentinel.App.Type
}

func LogBaseDir() string {
	return globalCfg.Sentinel.Log.Dir
}

func LogUsePid() bool {
	return globalCfg.Sentinel.Log.UsePid
}

func MetricLogFlushIntervalSec() uint32 {
	return globalCfg.Sentinel.Log.Metric.FlushIntervalSec
}

func MetricLogSingleFileMaxSize() uint64 {
	return globalCfg.Sentinel.Log.Metric.SingleFileMaxSize
}

func MetricLogMaxFileAmount() uint32 {
	return globalCfg.Sentinel.Log.Metric.MaxFileCount
}

func SystemStatCollectIntervalMs() uint32 {
	return globalCfg.Sentinel.Stat.System.CollectIntervalMs
}
