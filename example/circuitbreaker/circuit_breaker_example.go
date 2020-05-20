package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/util"
)

type stateChangeTestListener struct {
}

func (s *stateChangeTestListener) OnChangeToClosed(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("From %s to Closed, time: %d\n", prev.String(), util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnChangeToOpen(prev circuitbreaker.State, rule circuitbreaker.Rule, snapshot interface{}) {
	fmt.Printf("From %s to Open, snapshot: %.2f, time: %d\n", prev.String(), snapshot, util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnChangeToHalfOpen(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("From %s to Half-Open, time: %d\n", prev.String(), util.CurrentTimeMillis())
}

func main() {
	err := api.InitDefault()
	if err != nil {
		log.Fatal(err)
	}
	ch := make(chan struct{})
	circuitbreaker.RegisterStatusSwitchListeners(&stateChangeTestListener{})

	_, err = circuitbreaker.LoadRules([]circuitbreaker.Rule{
		circuitbreaker.NewSlowRtRule("abc", 20000, 5000, 50, 20, 0.5),
		circuitbreaker.NewErrorRatioRule("abc", 20000, 5000, 20, 0.5),
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			e, b := api.Entry("abc")
			if b != nil {
				fmt.Println("blocked")
				time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
			} else {
				fmt.Println("passed")
				time.Sleep(time.Duration(rand.Uint64()%80) * time.Millisecond)
				e.Exit()
			}
		}
	}()
	go func() {
		for {
			e, b := api.Entry("abc")
			if b != nil {
				fmt.Println("blocked")
				time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
			} else {
				fmt.Println("passed")
				time.Sleep(time.Duration(rand.Uint64()%80) * time.Millisecond)
				e.Exit()
			}
		}
	}()

	<-ch
}
