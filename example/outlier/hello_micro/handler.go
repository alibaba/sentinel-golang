package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	proto "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

var nodeCrash = false // Set the flag for whether to simulate node crash

type TestHandler struct {
	id        int
	startTime time.Time
}

func (s *TestHandler) Ping(ctx context.Context, req *proto.Request, rsp *proto.Response) error {
	rsp.Result = fmt.Sprintf("Welcome, I am node%d!", s.id)
	if nodeCrash {
		return nil
	}
	faultStartTime := s.startTime.Add(5 * time.Second).Add(time.Duration(s.id) * 5 * time.Second)
	faultEndTime := faultStartTime.Add(10 * time.Second)
	currentTime := time.Now()
	// If currentTime is in the time range of the business error
	if currentTime.After(faultStartTime) && currentTime.Before(faultEndTime) {
		return errors.New("internal server error")
	}
	return nil
}

func getIDWithAddress(address string) int {
	return int(address[len(address)-1])
}
