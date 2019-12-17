package main

import (
	"github.com/sentinel-group/sentinel-golang/core"
	"github.com/sentinel-group/sentinel-golang/util"
	"log"
	"time"
)

func main() {
	util.InitDefaultLoggerToConsole()
	log.Println("=================start=================")
	ctx := core.Entry("aaaaaa")
	time.Sleep(time.Second * 1)
	log.Println("Call service")
	ctx.Exit()
	log.Println("=================end=================")
}
