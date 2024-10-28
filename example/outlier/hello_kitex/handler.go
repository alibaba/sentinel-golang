package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
)

type HelloImpl struct {
	id        int
	startTime time.Time
}

// If the simulated node crashes, then Echo returns correct directly, otherwise Echo needs to simulate a business error
func (s *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	message := fmt.Sprintf("Welcome %s,I am node%d", req.Message, s.id)
	if *nodeCrashFlag {
		return &api.Response{Message: message}, nil
	}
	faultStartTime := s.startTime.Add(5 * time.Second).Add(time.Duration(s.id) * 5 * time.Second)
	faultEndTime := faultStartTime.Add(20 * time.Second)
	currentTime := time.Now()
	// If currentTime is in the time range of the business error
	if currentTime.After(faultStartTime) && currentTime.Before(faultEndTime) {
		return nil, errors.New("internal server error")
	}
	return &api.Response{Message: message}, nil
}

func getIDWithAddress(address string) int {
	return int(address[len(address)-1] - '0')
}
