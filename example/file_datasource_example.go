package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/alibaba/sentinel-golang/extension/datasource/file"
	"github.com/alibaba/sentinel-golang/util"

	// import flow
	_ "github.com/alibaba/sentinel-golang/core/flow"
)
// flow.GetRules() = [{"resource":"some-test","limitApp":"default","grade":1,"count":10,"strategy":0,"controlBehavior":0,"warmUpPeriodSec":0,"maxQueueingTimeMs":0,"clusterMode":false,"clusterConfig":{"thresholdType":0}}]

func init() {
	// note(gorexlv): just for testing, you'd better put those config into config file
	var datasourceConfig = `
[
	{
		"resource": "some-test",
		"grade": 1,
		"count":5,
		"controlBehavior":0
	}
]
`
	if err := ioutil.WriteFile(filePath, []byte(datasourceConfig), os.ModePerm); err != nil {
		panic(err)
	}
}

var filePath = os.TempDir() + "file_datasource.json"

func main() {
	ds := file.New(filePath)

	if err := ds.ReadConfig(); err != nil {
		panic(err)
	}

	// We should initialize Sentinel first.
	err := sentinel.InitDefault()
	if err != nil {
		log.Fatalf("Unexpected error: %+v", err)
	}

	fmt.Printf("flow.GetRules() = %+v\n", flow.GetRules())

	ch := make(chan struct{})

	go func() {
		time.Sleep(time.Second*10)
		_ = os.Remove(filePath)
		time.Sleep(time.Second)
		fmt.Printf(" ===> flow.GetRules() = %+v\n", flow.GetRules())

		time.Sleep(time.Second * 1)
		os.Exit(1)
	}()

	for i := 0; i < 10; i++ {
		go func() {
			for {
				e, b := sentinel.Entry("some-test", sentinel.WithTrafficType(base.Inbound))
				if b != nil {
					// Blocked. We could get the block reason from the BlockError.
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)
				} else {
					// Passed, wrap the logic here.
					fmt.Println(util.CurrentTimeMillis(), "passed")
					time.Sleep(time.Duration(rand.Uint64()%10) * time.Millisecond)

					// Be sure the entry is exited finally.
					e.Exit()
				}

			}
		}()
	}
	<-ch
}
