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
	FatalLevel
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

	Error(msg string, keysAndValues ...interface{})

	Fatal(msg string, keysAndValues ...interface{})
}

// sentinel general logger
type DefaultLogger struct {
	// entity to log
	log *log.Logger
}

func handleKV(k, v interface{}) (string, string) {
	if kStr, isStr := k.(string); !isStr {
		return fmt.Sprintf("%+v", k), fmt.Sprintf("%+v", v)
	} else {
		return kStr, fmt.Sprintf("%+v", v)
	}
}

func caller(depth int) (file string, line int) {
	_, file, line, ok := runtime.Caller(depth)
	if !ok {
		file = "???"
		line = 0
	}
	return
}

func AssembleMsg(depth int, logLevel, msg string, keysAndValues ...interface{}) string {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("%s    ", time.Now().Format("2006-01-02T15:04:05.000")))
	file, line := caller(depth)
	sb.WriteString(fmt.Sprintf("%s:%d    %s    %s    ", file, line, logLevel, msg))

	if len(keysAndValues) == 0 {
		return sb.String()
	}
	// length of kvs is odd
	if len(keysAndValues)&1 != 0 {
		sb.WriteString(fmt.Sprintf("%+v", keysAndValues))
		return sb.String()
	}

	// length of kvs is even
	kvsMap := make(map[string]string, 0)
	for i := 0; i < len(keysAndValues); {
		k := keysAndValues[i]
		v := keysAndValues[i+1]
		kStr, vStr := handleKV(k, v)
		kvsMap[kStr] = vStr
		i = i + 2
	}
	msgs, err := json.Marshal(kvsMap)
	if err != nil {
		sb.WriteString(fmt.Sprintf("%+v", keysAndValues))
		return sb.String()
	}
	sb.Write(msgs)
	return sb.String()
}

func (l *DefaultLogger) Debug(msg string, keysAndValues ...interface{}) {
	if DebugLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(globalCallerDepth, "DEBUG", msg, keysAndValues...))
}

func (l *DefaultLogger) Info(msg string, keysAndValues ...interface{}) {
	if InfoLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(globalCallerDepth, "INFO", msg, keysAndValues...))
}

func (l *DefaultLogger) Warn(msg string, keysAndValues ...interface{}) {
	if WarnLevel < globalLogLevel {
		return
	}

	l.log.Print(AssembleMsg(globalCallerDepth, "WARNING", msg, keysAndValues...))
}

func (l *DefaultLogger) Error(msg string, keysAndValues ...interface{}) {
	if ErrorLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(globalCallerDepth, "ERROR", msg, keysAndValues...))
}

func (l *DefaultLogger) Fatal(msg string, keysAndValues ...interface{}) {
	if FatalLevel < globalLogLevel {
		return
	}
	l.log.Print(AssembleMsg(globalCallerDepth, "FATAL", msg, keysAndValues...))
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

func Error(msg string, keysAndValues ...interface{}) {
	globalLogger.Error(msg, keysAndValues...)
}

func Fatal(msg string, keysAndValues ...interface{}) {
	globalLogger.Fatal(msg, keysAndValues...)
}
