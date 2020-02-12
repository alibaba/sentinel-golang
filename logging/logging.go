package logging

import (
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

var (
	globalLogLevel = Info

	defaultLogger = NewConsoleLogger(defaultNamespace)
)

func GetGlobalLoggerLevel() Level {
	return globalLogLevel
}

func SetGlobalLoggerLevel(l Level) {
	globalLogLevel = l
}

func GetDefaultLogger() Logger {
	return defaultLogger
}

func NewConsoleLogger(namespace string) *SentinelLogger {
	return &SentinelLogger{
		log:       log.New(os.Stdout, "", log.LstdFlags),
		namespace: defaultNamespace,
	}
}

// outputFile is the full path(absolute path)
func NewSimpleFileLogger(filepath, namespace string, flag int) (*SentinelLogger, error) {
	logFile, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return nil, err
	}
	return &SentinelLogger{
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
type SentinelLogger struct {
	// entity to log
	log *log.Logger
	// namespace
	namespace string
}

func merge(namespace, logLevel, msg string) string {
	return fmt.Sprintf("[%s] [%s] %s", namespace, logLevel, msg)
}

func (l *SentinelLogger) Debug(v ...interface{}) {
	if Debug < globalLogLevel || len(v) == 0 {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Debugf(format string, v ...interface{}) {
	if Debug < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Info(v ...interface{}) {
	if Info < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Infof(format string, v ...interface{}) {
	if Info < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Warn(v ...interface{}) {
	if Warn < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Warnf(format string, v ...interface{}) {
	if Warn < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Error(v ...interface{}) {
	if Error < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Errorf(format string, v ...interface{}) {
	if Error < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Fatal(v ...interface{}) {
	if Fatal < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Fatalf(format string, v ...interface{}) {
	if Fatal < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Panic(v ...interface{}) {
	if Panic < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Panicf(format string, v ...interface{}) {
	if Panic < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprintf(format, v...)))
}
