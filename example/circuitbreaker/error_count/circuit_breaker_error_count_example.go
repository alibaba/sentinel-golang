package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/circuitbreaker"
	"github.com/alibaba/sentinel-golang/core/config"
	"github.com/alibaba/sentinel-golang/logging"
	"github.com/alibaba/sentinel-golang/util"
)

type stateChangeTestListener struct {
}

func (s *stateChangeTestListener) OnTransformToClosed(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	logging.Warn(fmt.Sprintf("circuit breaker status changes from %s to Closed.", prev.String()), "timestamp", util.CurrentTimeMillis(),
		"strategy", rule.Strategy.String(), "rule", rule)
}

func (s *stateChangeTestListener) OnTransformToOpen(prev circuitbreaker.State, rule circuitbreaker.Rule, snapshot interface{}) {
	logging.Warn(fmt.Sprintf("circuit breaker status changes from %s to Open.", prev.String()), "timestamp", util.CurrentTimeMillis(),
		"strategy", rule.Strategy.String(), "snapshot", snapshot, "rule", rule)
}

func (s *stateChangeTestListener) OnTransformToHalfOpen(prev circuitbreaker.State, rule circuitbreaker.Rule) {
	logging.Warn(fmt.Sprintf("circuit breaker status changes from %s to HalfOpen.", prev.String()), "timestamp", util.CurrentTimeMillis(), "strategy",
		rule.Strategy.String(), "rule", rule)
}

func main() {
	conf := config.NewDefaultConfig()
	conf.Sentinel.Log.Logger = logging.NewConsoleLogger()
	conf.Sentinel.Stat.System.CollectIntervalMs = 0
	err := sentinel.InitWithConfig(conf)
	if err != nil {
		log.Fatal(err)
	}
	ch := make(chan struct{})
	// Register a state change listener so that we could observer the state change of the internal circuit breaker.
	circuitbreaker.RegisterStateChangeListeners(&stateChangeTestListener{})

	_, err = circuitbreaker.LoadRules([]*circuitbreaker.Rule{
		// Statistic time span=10s, recoveryTimeout=3s, maxErrorCount=50
		{
			Resource:         "abc",
			Strategy:         circuitbreaker.ErrorCount,
			RetryTimeoutMs:   3000,
			MinRequestAmount: 10,
			StatIntervalMs:   5000,
			Threshold:        50,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	logging.Info("Sentinel Go circuit breaking demo is running. You may see the pass/block metric in the metric log.")
	go func() {
		for {
			e, b := sentinel.Entry("abc")
			if b != nil {
				// g1 blocked
				time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
			} else {
				if rand.Uint64()%20 > 9 {
					// Record current invocation as error.
					sentinel.TraceError(e, errors.New("biz error"))
				}
				// g1 passed
				time.Sleep(time.Duration(rand.Uint64()%80+10) * time.Millisecond)
				e.Exit()
			}
		}
	}()
	go func() {
		for {
			e, b := sentinel.Entry("abc")
			if b != nil {
				// g2 blocked
				time.Sleep(time.Duration(rand.Uint64()%20) * time.Millisecond)
			} else {
				// g2 passed
				time.Sleep(time.Duration(rand.Uint64()%80) * time.Millisecond)
				e.Exit()
			}
		}
	}()
	<-ch
}
