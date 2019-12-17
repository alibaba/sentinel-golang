package util

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// log level: PANIC、FATAL、ERROR、WARN、INFO、DEBUG
type LogLevel uint8

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
	FATAL
	PANIC
)

// format default console log
func InitDefaultLoggerToConsole() {
	fmt.Println("Init default log, output to console")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[sentinel]")
}

var (
	GlobalLogLevel = DEBUG
	defaultLogger  *SentinelLogger
	// init default logger only once
	initLogger sync.Once
)

func SetGlobalLoggerLevel(l LogLevel) {
	GlobalLogLevel = l
}

func init() {
	fmt.Println("util logging init")
	initLogger.Do(func() {
		defaultLogger = NewSentinelFileLogger("sentinel.log", "default", log.LstdFlags|log.Lshortfile)
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
		log.Fatal("open log file error, ", err)
	}
	return &SentinelLogger{
		log:       log.New(logFile, "", flag),
		namespace: namespace,
	}
}

type Logger interface {
	Debug(v ...interface{})
	Debugln(v ...interface{})
	Debugf(format string, v ...interface{})

	Info(v ...interface{})
	Infoln(v ...interface{})
	Infof(format string, v ...interface{})

	Warning(v ...interface{})
	Warningln(v ...interface{})
	Warningf(format string, v ...interface{})

	Error(v ...interface{})
	Errorln(v ...interface{})
	Errorf(format string, v ...interface{})

	Fatal(v ...interface{})
	Fatalln(v ...interface{})
	Fatalf(format string, v ...interface{})

	Panic(v ...interface{})
	Panicln(v ...interface{})
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
	if DEBUG < GlobalLogLevel || len(v) == 0 {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Debugln(v ...interface{}) {
	if DEBUG < GlobalLogLevel {
		return
	}
	l.log.Println(merge(l.namespace, "DEBUG", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Debugf(format string, v ...interface{}) {
	if DEBUG < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "DEBUG", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Info(v ...interface{}) {
	if INFO < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Infoln(v ...interface{}) {
	if INFO < GlobalLogLevel {
		return
	}
	l.log.Println(merge(l.namespace, "INFO", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Infof(format string, v ...interface{}) {
	if INFO < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "INFO", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Warning(v ...interface{}) {
	if WARNING < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Warningln(v ...interface{}) {
	if WARNING < GlobalLogLevel {
		return
	}
	l.log.Println(merge(l.namespace, "WARNING", fmt.Sprint(v...)))
}

func (l *SentinelLogger) Warningf(format string, v ...interface{}) {
	if WARNING < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "WARNING", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Error(v ...interface{}) {
	if ERROR < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprint(v...)))
}
func (l *SentinelLogger) Errorln(v ...interface{}) {
	if ERROR < GlobalLogLevel {
		return
	}
	l.log.Println(merge(l.namespace, "ERROR", fmt.Sprint(v...)))
}
func (l *SentinelLogger) Errorf(format string, v ...interface{}) {
	if ERROR < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "ERROR", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Fatal(v ...interface{}) {
	if FATAL < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprint(v...)))
}
func (l *SentinelLogger) Fatalln(v ...interface{}) {
	if FATAL < GlobalLogLevel {
		return
	}
	l.log.Println(merge(l.namespace, "FATAL", fmt.Sprint(v...)))
}
func (l *SentinelLogger) Fatalf(format string, v ...interface{}) {
	if FATAL < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "FATAL", fmt.Sprintf(format, v...)))
}

func (l *SentinelLogger) Panic(v ...interface{}) {
	if PANIC < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprint(v...)))
}
func (l *SentinelLogger) Panicln(v ...interface{}) {
	if PANIC < GlobalLogLevel {
		return
	}
	l.log.Println(merge(l.namespace, "PANIC", fmt.Sprint(v...)))
}
func (l *SentinelLogger) Panicf(format string, v ...interface{}) {
	if PANIC < GlobalLogLevel {
		return
	}
	l.log.Print(merge(l.namespace, "PANIC", fmt.Sprintf(format, v...)))
}
