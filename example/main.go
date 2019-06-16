package main

import (
	"fmt"
	"github.com/sentinel-group/sentinel-golang/core"
	"math/rand"
	"sync"
	"time"
)

func main() {
	fmt.Println("=================start=================")
	wg := &sync.WaitGroup{}
	wg.Add(10)

	for i := 0; i < 10; i++ {
		test(wg)
	}
	wg.Wait()
	fmt.Println("=================end=================")
}

func test(wg *sync.WaitGroup) {
	rand.Seed(1000)
	r := rand.Int63() % 10
	time.Sleep(time.Duration(r) * time.Millisecond)
	_ = core.Entry("test")
	time.Sleep(time.Duration(r) * time.Millisecond)
	core.Exit("test")
	wg.Done()
}
