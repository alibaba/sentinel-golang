package core

import (
	"fmt"
	"log"
	"os"
)

func InitDefaultLoggerToConsole() {
	fmt.Println("Init default log, output to console")
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix("[sentinel]")
}

// outputFile is the full path(absolute path)
func NewFileLogger(outputFile, prefix string, flag int) *log.Logger {
	// get file info
	var logFile *os.File
	_, err := os.Stat(outputFile)
	if err == nil {
		logFile, err = os.Open(outputFile)
		if err != nil {
			log.Fatal("open log file error, ", err)
		}
	} else if err != nil && os.IsNotExist(err) {
		logFile, err = os.Create(outputFile)
		if err != nil {
			log.Fatal("create log file error, ", err)
		}
	} else {
		log.Fatal("open log file error, ", err)
	}
	return log.New(logFile, prefix, flag)
}
