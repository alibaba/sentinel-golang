package logging

import (
	"errors"
	"fmt"
	"log"
	"os"
)

// Level represents the level of logging.
type Level uint8

const (
	Debug Level = iota
	Info
	Warn
	Error
	Fatal
	Panic
)

const (
	DefaultNamespace = "default"
	// RecordLogFileName represents the default file name of the record log.
	RecordLogFileName = "sentinel-record.log"
	DefaultDirName    = "logs" + string(os.PathSeparator) + "csp" + string(os.PathSeparator)
)

var (
	globalLogLevel = Info

	globalLogger = NewConsoleLogger(DefaultNamespace)
)

func GetGlobalLoggerLevel() Level {
	return globalLogLevel
}

func SetGlobalLoggerLevel(l Level) {
	globalLogLevel = l
}

func GetGlobalLogger() Logger {
	return globalLogger
}

// Note: Not thread-safe
func ResetGlobalLogger(log Logger) error {
	if log == nil {
		return errors.New("nil logger")
	}
	globalLogger = log
	return nil
}

func NewConsoleLogger(namespace string) Logger {
	return &DefaultLogger{
		log:       log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
		namespace: namespace,
	}
}

// outputFile is the full path(absolute path)
func NewSimpleFileLogger(filepath, namespace string, flag int) (Logger, error) {
	logFile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &DefaultLogger{
		log:       log.New(logFile, "", flag),
		namespace: namespace,
	}, err
}

type Logger interface {
	Debug(v ...interface{})
	Debugf(format string, v ...interface{})

	Info(v ...interface{})
	Infof(format string, v ...interface{})

	Warn(v ...interface{})
	Warnf(format string, v ...interface{})

	Error(v ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicf(format string, v ...interface{})
}

// sentinel general logger
type DefaultLogger struct {
	// entity to log
	log *log.Logger
	// namespace
	namespace string
}

func merge(namespace, logLevel, msg string) string {
	return fmt.Sprintf("[%s] [%s] %s", namespace, logLevel, msg)
}

func (l *DefaultLogger) Debug(v ...interface{}) {
	if Debug < globalLogLevel || len(v) == 0 {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	if Debug < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Info(v ...interface{}) {
	if Info < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	if Info < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Warn(v ...interface{}) {
	if Warn < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	if Warn < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Error(v ...interface{}) {
	if Error < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	if Error < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Fatal(v ...interface{}) {
	if Fatal < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	if Fatal < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Panic(v ...interface{}) {
	if Panic < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Panicf(format string, v ...interface{}) {
	if Panic < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprintf(format, v...)))
}
