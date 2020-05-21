package main

import (
	"errors"
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
	fmt.Printf("rule.steategy: %+v, From %s to Closed, time: %d\n", rule.BreakerStrategy(), prev.String(), util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnChangeToOpen(prev circuitbreaker.State, rule circuitbreaker.Rule, snapshot interface{}) {
	fmt.Printf("rule.steategy: %+v, From %s to Open, snapshot: %.2f, time: %d\n", rule.BreakerStrategy(), prev.String(), snapshot, util.CurrentTimeMillis())
}

func (s *stateChangeTestListener) OnChangeToHalfOpen(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	fmt.Printf("rule.steategy: %+v, From %s to Half-Open, time: %d\n", rule.BreakerStrategy(), prev.String(), util.CurrentTimeMillis())
}

func main() {
	err := api.InitDefault()
	if err != nil {
		log.Fatal(err)
	}
	ch := make(chan struct{})
	circuitbreaker.RegisterStatusSwitchListeners(&stateChangeTestListener{})

	_, err = circuitbreaker.LoadRules([]circuitbreaker.Rule{
		circuitbreaker.NewSlowRtRule("abc", 10000, 3000, 50, 10, 0.5),
		circuitbreaker.NewErrorRatioRule("abc", 10000, 3000, 10, 0.5),
	})
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			e, b := api.Entry("abc")
			if b != nil {
				fmt.Println("g1blocked")
				time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
			} else {
				if rand.Uint64()%20 > 9 {
					e.SetError(errors.New("biz error"))
				}
				fmt.Println("g1passed")
				time.Sleep(time.Duration(rand.Uint64()%80) * time.Millisecond)
				e.Exit()
			}
		}
	}()
	go func() {
		for {
			e, b := api.Entry("abc")
			if b != nil {
				fmt.Println("g2blocked")
				time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
			} else {
				fmt.Println("g2passed")
				time.Sleep(time.Duration(rand.Uint64()%80) * time.Millisecond)
				e.Exit()
			}
		}
	}()

	<-ch
}
