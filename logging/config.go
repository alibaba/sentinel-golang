package logging

import (
	"github.com/pkg/errors"
	"github.com/sentinel-group/sentinel-golang/core/config"
	"os"
	"strings"
	"sync"
)

const defaultDirName = "logs" + string(os.PathSeparator) + "csp"

var (
	logBaseDir     string
	logNameWithPid = false

	initLoggerConfig sync.Once
)

func InitializeLogConfigFromEnv() error {
	addPid := os.Getenv(config.LogNamePidEnvKey)
	if strings.ToLower(addPid) == "true" {
		logNameWithPid = true
	}
	logDir := os.Getenv(config.LogDirEnvKey)
	if logDir == "" {
		d, err := getDefaultLogDir()
		if err != nil {
			return err
		}
		logDir = d
	}
	logBaseDir = logDir

	// TODO: create dir if not exists

	return nil
}

func InitializeLogConfig(logPath string, usePid bool) error {
	if logPath == "" {
		return errors.New("invalid empty log path")
	}
	logBaseDir = logPath
	logNameWithPid = usePid
	return nil
}

func getDefaultLogDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return addSeparatorIfNeeded(home) + defaultDirName, nil
}

func addSeparatorIfNeeded(path string) string {
	s := string(os.PathSeparator)
	if !strings.HasSuffix(path, s) {
		return path + s
	}
	return path
}
