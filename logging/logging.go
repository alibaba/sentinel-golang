package logging

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// log level: PANIC、FATAL、ERROR、WARN、INFO、DEBUG
type Level uint8

const (
	Debug Level = iota
	Info
	Warn
	Error
	Fatal
	Panic
)

const RecordLogFileName = "sentinel-record.log"

// format default console log
func InitDefaultLoggerToConsole() {
	fmt.Println("Init default log, output to console")
	log.SetFlags(log.LstdFlags)
	log.SetPrefix("[sentinel]")
}

var (
	GlobalLogLevel = Debug
	defaultLogger  *SentinelLogger
	// init default logger only once
	initLogger sync.Once
)

func SetGlobalLoggerLevel(l Level) {
	GlobalLogLevel = l
}

func init() {
	initLogger.Do(func() {
		defaultLogger = NewSentinelFileLogger(RecordLogFileName, "default", log.LstdFlags)
	})
}

func GetDefaultLogger() *SentinelLogger {
	return defaultLogger
}

// outputFile is the full path(absolute path)
func NewSentinelFileLogger(outputFile, namespace string, flag int) *SentinelLogger {
	//get file info
	var logFile *os.File
	logFile, err := os.OpenFile(outputFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		log.Fatal("open log file error:", err)
	}
	return &SentinelLogger{
		log:       log.New(logFile, "", flag),
		namespace: namespace,
	}
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
	if Debug < GlobalLogLevel || len(v) == 0 {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Debugf(format string, v ...interface{}) {
	if Debug < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Info(v ...interface{}) {
	if Info < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Infof(format string, v ...interface{}) {
	if Info < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Warn(v ...interface{}) {
	if Warn < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Warnf(format string, v ...interface{}) {
	if Warn < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Error(v ...interface{}) {
	if Error < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Errorf(format string, v ...interface{}) {
	if Error < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Fatal(v ...interface{}) {
	if Fatal < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Fatalf(format string, v ...interface{}) {
	if Fatal < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Panic(v ...interface{}) {
	if Panic < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Panicf(format string, v ...interface{}) {
	if Panic < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprintf(format, v...)))
}
