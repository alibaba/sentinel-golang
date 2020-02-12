package logging

import (
	"fmt"
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

const (
	LogDirEnvKey     = "SENTINEL_LOG_DIR"
	LogNamePidEnvKey = "SENTINEL_LOG_USE_PID"

	// RecordLogFileName represents the default file name of the record log.
	RecordLogFileName = "sentinel-record.log"

	defaultNamespace = "default"
	defaultDirName   = "logs" + string(os.PathSeparator) + "csp" + string(os.PathSeparator)
)

var (
	logBaseDir     string
	logNameWithPid = false

	logInitialized int32 = 0
)

// LogBaseDir returns the directory of Sentinel logs.
func LogBaseDir() string {
	return logBaseDir
}

// LogNameWithPid indicates whether log filename ends with the process ID (PID).
func LogNameWithPid() bool {
	return logNameWithPid
}

// InitializeLogConfigFromEnv initializes the configuration from system environment.
// If relevant properties are absent, Sentinel will use the default configuration.
func InitializeLogConfigFromEnv() error {
	addPid := os.Getenv(LogNamePidEnvKey)
	if strings.ToLower(addPid) == "true" {
		logNameWithPid = true
	}
	logDir := os.Getenv(LogDirEnvKey)
	if logDir == "" {
		d, err := getDefaultLogDir()
		if err != nil {
			return err
		}
		logDir = d
	}
	logBaseDir = logDir

	return InitializeLogConfig(logBaseDir, logNameWithPid)
}

func InitializeLogConfig(logPath string, usePid bool) error {
	if logPath == "" {
		return errors.New("invalid empty log path")
	}
	if !atomic.CompareAndSwapInt32(&logInitialized, 0, 1) {
		return nil
	}

	logBaseDir = logPath
	logNameWithPid = usePid

	if err := createDirIfNotExists(logBaseDir); err != nil {
		return err
	}
	return reconfigureRecordLogger()
}

func reconfigureRecordLogger() error {
	logDir := addSeparatorIfNeeded(LogBaseDir())
	filePath := logDir + RecordLogFileName
	if logNameWithPid {
		filePath = filePath + ".pid" + strconv.Itoa(os.Getpid())
	}
	if defaultLogger == nil {
		return errors.New("Unexpected state: defaultLogger == nil")
	}
	logFile, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}

	// Note: not thread-safe!
	defaultLogger.log = log.New(logFile, "", log.LstdFlags)
	defaultLogger.namespace = defaultNamespace

	fmt.Println("INFO: log base directory is: " + logDir)

	return nil
}

func createDirIfNotExists(dirname string) error {
	if _, err := os.Stat(dirname); err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(dirname, os.ModePerm)
		} else {
			return err
		}
	}
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
