package logging

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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
	DefaultDirName    = "logs" + string(os.PathSeparator) + "csp" + string(os.PathSeparator)
)

var (
	globalLogLevel    = InfoLevel
	globalCallerDepth = 4
	globalLogger      = NewConsoleLogger()
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
	return
}

func assembleMsg(depth int, logLevel, msg string, err error, keysAndValues ...interface{}) string {
	type logEntity struct {
		Timestamp string                 `json:"timestamp"`
		Caller    string                 `json:"caller"`
		LogLevel  string                 `json:"logLevel"`
		Msg       string                 `json:"msg"`
		Params    map[string]interface{} `json:"params"`
	}
	file, line := caller(depth)
	timeStr := time.Now().Format("2006-01-02 15:04:05.520")
	caller := fmt.Sprintf("%s:%d", file, line)
	logE := &logEntity{
		Timestamp: timeStr,
		Caller:    caller,
		LogLevel:  logLevel,
		Msg:       msg,
		Params:    make(map[string]interface{}),
	}

	if len(keysAndValues)&1 != 0 {
		logE.Params["params"] = fmt.Sprintf("%+v", keysAndValues)
	} else if len(keysAndValues) != 0 {
		for i := 0; i < len(keysAndValues); {
			k := keysAndValues[i]
			v := keysAndValues[i+1]
			kStr, kIsStr := k.(string)
			if !kIsStr {
				logE.Params[fmt.Sprintf("%+v", k)] = v
			} else {
				logE.Params[kStr] = v
			}
			i = i + 2
		}
	}
	msgs, e := json.Marshal(logE)
	if e != nil {
		return fmt.Sprintf("%s %s %s %s %+v %+v", timeStr, caller, logLevel, msg, e, keysAndValues)
	}
	sb := strings.Builder{}
	sb.Write(msgs)
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
	l.log.Print(assembleMsg(globalCallerDepth, "DEBUG", msg, nil, keysAndValues...))
}

func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	if InfoLevel < globalLogLevel {
		return
	}
	l.log.Print(assembleMsg(globalCallerDepth, "INFO", msg, nil, keysAndValues...))
}

func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	if WarnLevel < globalLogLevel {
		return
	}

	l.log.Print(assembleMsg(globalCallerDepth, "WARNING", msg, nil, keysAndValues...))
}

func (l *DefaultLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	if ErrorLevel < globalLogLevel {
		return
	}
	l.log.Print(assembleMsg(globalCallerDepth, "ERROR", msg, err, keysAndValues...))
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
