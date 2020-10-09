package logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// Level represents the level of logging.
type Level uint8

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

const (
	// RecordLogFileName represents the default file name of the record log.
	RecordLogFileName = "sentinel-record.log"
	GlobalCallerDepth = 4
)

var (
	DefaultDirName = filepath.Join("logs", "csp")
)

var (
	globalLogLevel = InfoLevel
	globalLogger   = NewConsoleLogger()

	FrequentErrorOnce = &sync.Once{}
)

func GetGlobalLoggerLevel() Level {
	return globalLogLevel
}

func SetGlobalLoggerLevel(l Level) {
	globalLogLevel = l
}

// Note: Not thread-safe
func ResetGlobalLogger(log Logger) error {
	if log == nil {
		return errors.New("nil logger")
	}
	globalLogger = log
	return nil
}

func NewConsoleLogger() Logger {
	return &DefaultLogger{
		log: log.New(os.Stdout, "", 0),
	}
}

// outputFile is the full path(absolute path)
func NewSimpleFileLogger(filepath string) (Logger, error) {
	logFile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &DefaultLogger{
		log: log.New(logFile, "", 0),
	}, nil
}

type Logger interface {
	Debug(msg string, keysAndValues ...interface{})

	// Info logs a non-error message with the given key/value pairs as context.
	//
	// The msg argument should be used to add some constant description to
	// the log line.  The key/value pairs can then be used to add additional
	// variable information.  The key/value pairs should alternate string
	// keys and arbitrary values.
	Info(msg string, keysAndValues ...interface{})

	Warn(msg string, keysAndValues ...interface{})

	Error(err error, msg string, keysAndValues ...interface{})
}

// sentinel general logger
type DefaultLogger struct {
	log *log.Logger
}

func caller(depth int) (file string, line int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}

	// extract
	if osType := runtime.GOOS; osType == "windows" {
		file = strings.ReplaceAll(file, "\\", "/")
	}
	idx := strings.LastIndex(file, "/")
	file = file[idx+1:]
	return
}

func AssembleMsg(depth int, logLevel, msg string, err error, keysAndValues ...interface{}) string {
	sb := strings.Builder{}
	file, line := caller(depth)
	timeStr := time.Now().Format("2006-01-02 15:04:05.520")
	caller := fmt.Sprintf("%s:%d", file, line)
	sb.WriteString("{")

	sb.WriteByte('"')
	sb.WriteString("timestamp")
	sb.WriteByte('"')
	sb.WriteByte(':')
	sb.WriteByte('"')
	sb.WriteString(timeStr)
	sb.WriteByte('"')
	sb.WriteByte(',')

	sb.WriteByte('"')
	sb.WriteString("caller")
	sb.WriteByte('"')
	sb.WriteByte(':')
	sb.WriteByte('"')
	sb.WriteString(caller)
	sb.WriteByte('"')
	sb.WriteByte(',')

	sb.WriteByte('"')
	sb.WriteString("logLevel")
	sb.WriteByte('"')
	sb.WriteByte(':')
	sb.WriteByte('"')
	sb.WriteString(logLevel)
	sb.WriteByte('"')
	sb.WriteByte(',')

	sb.WriteByte('"')
	sb.WriteString("msg")
	sb.WriteByte('"')
	sb.WriteByte(':')
	sb.WriteByte('"')
	sb.WriteString(msg)
	sb.WriteByte('"')

	kvLen := len(keysAndValues)
	if kvLen&1 != 0 {
		sb.WriteByte(',')
		sb.WriteByte('"')
		sb.WriteString("kvs")
		sb.WriteByte('"')
		sb.WriteByte(':')
		sb.WriteByte('"')
		sb.WriteString(fmt.Sprintf("%+v", keysAndValues))
		sb.WriteByte('"')
	} else if kvLen != 0 {
		for i := 0; i < kvLen; {
			k := keysAndValues[i]
			v := keysAndValues[i+1]
			kStr, kIsStr := k.(string)
			if !kIsStr {
				kStr = fmt.Sprintf("%+v", k)
			}
			sb.WriteByte(',')
			sb.WriteByte('"')
			sb.WriteString(kStr)
			sb.WriteByte('"')
			sb.WriteByte(':')
			vStr, vIsStr := v.(string)
			if !vIsStr {
				if vbs, err := json.Marshal(v); err != nil {
					sb.WriteByte('"')
					sb.WriteString(fmt.Sprintf("%+v", v))
					sb.WriteByte('"')
				} else {
					sb.WriteString(string(vbs))
				}
			} else {
				sb.WriteByte('"')
				sb.WriteString(vStr)
				sb.WriteByte('"')
			}
			i = i + 2
		}
	}
	sb.WriteByte('}')
	if err != nil {
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%+v", err))
	}
	return sb.String()
}

func (l *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	if DebugLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(GlobalCallerDepth, "DEBUG", msg, nil, keysAndValues...))
}

func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	if InfoLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(GlobalCallerDepth, "INFO", msg, nil, keysAndValues...))
}

func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	if WarnLevel < globalLogLevel {
		return
	}

	l.log.Print(AssembleMsg(GlobalCallerDepth, "WARNING", msg, nil, keysAndValues...))
}

func (l *DefaultLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if ErrorLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(GlobalCallerDepth, "ERROR", msg, err, keysAndValues...))
}

func Debug(msg string, keysAndValues ...interface{}) {
	globalLogger.Debug(msg, keysAndValues...)
}

func Info(msg string, keysAndValues ...interface{}) {
	globalLogger.Info(msg, keysAndValues...)
}

func Warn(msg string, keysAndValues ...interface{}) {
	globalLogger.Warn(msg, keysAndValues...)
}

func Error(err error, msg string, keysAndValues ...interface{}) {
	globalLogger.Error(err, msg, keysAndValues...)
}
