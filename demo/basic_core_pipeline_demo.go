package main

import (
	"github.com/sentinel-group/sentinel-golang/core"
	"log"
	"time"
)

func main() {
	core.InitDefaultLoggerToConsole()
	log.Println("=================start=================")
	ctx := core.Entry("aaaaaa")
	time.Sleep(time.Second * 1)
	log.Println("Call service")
	ctx.Exit()
	log.Println("=================end=================")
}
