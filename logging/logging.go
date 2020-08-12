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
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

const (
	DefaultNamespace = "default"
	// RecordLogFileName represents the default file name of the record log.
	RecordLogFileName = "sentinel-record.log"
	DefaultDirName    = "logs" + string(os.PathSeparator) + "csp" + string(os.PathSeparator)
)

var (
	globalLogLevel = InfoLevel
	globalLogger   = NewConsoleLogger(DefaultNamespace)
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
	if DebugLevel < globalLogLevel || len(v) == 0 {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Debugf(format string, v ...interface{}) {
	if DebugLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Info(v ...interface{}) {
	if InfoLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Infof(format string, v ...interface{}) {
	if InfoLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Warn(v ...interface{}) {
	if WarnLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Warnf(format string, v ...interface{}) {
	if WarnLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Error(v ...interface{}) {
	if ErrorLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Errorf(format string, v ...interface{}) {
	if ErrorLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Fatal(v ...interface{}) {
	if FatalLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Fatalf(format string, v ...interface{}) {
	if FatalLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprintf(format, v...)))
}

func (l *DefaultLogger) Panic(v ...interface{}) {
	if PanicLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprint(v...)))
}

func (l *DefaultLogger) Panicf(format string, v ...interface{}) {
	if PanicLevel < globalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprintf(format, v...)))
}

func Debug(v ...interface{}) {
	globalLogger.Debug(v...)
}

func Debugf(format string, v ...interface{}) {
	globalLogger.Debugf(format, v...)
}

func Info(v ...interface{}) {
	globalLogger.Info(v...)
}

func Infof(format string, v ...interface{}) {
	globalLogger.Infof(format, v...)
}

func Warn(v ...interface{}) {
	globalLogger.Warn(v...)
}

func Warnf(format string, v ...interface{}) {
	globalLogger.Warnf(format, v...)
}

func Error(v ...interface{}) {
	globalLogger.Error(v...)
}

func Errorf(format string, v ...interface{}) {
	globalLogger.Errorf(format, v...)
}

func Fatal(v ...interface{}) {
	globalLogger.Fatal(v...)
}

func Fatalf(format string, v ...interface{}) {
	globalLogger.Fatalf(format, v...)
}

func Panic(v ...interface{}) {
	globalLogger.Panic(v...)
}

func Panicf(format string, v ...interface{}) {
	globalLogger.Panicf(format, v...)
}
