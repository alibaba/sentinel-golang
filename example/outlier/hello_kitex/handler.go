package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/kitex-examples/hello/kitex_gen/api"
)

// HelloImpl implements the last service interface defined in the IDL.
type HelloImpl struct {
	address    string
	networkErr bool
	isCrash    bool
	done       chan struct{}
}

func NewHello() api.Hello {
	res := &HelloImpl{address: *addressFlag, networkErr: *errorFlag}
	res.done = make(chan struct{})
	go func() {
		start := 10 * time.Second
		end := 15 * time.Second
		timer1 := time.NewTimer(start)
		timer2 := time.NewTimer(end)

		<-timer1.C
		res.isCrash = true

		<-timer2.C
		res.isCrash = false
		res.done <- struct{}{}
	}()
	return res
}

// Echo implements the HelloImpl interface.
func (s *HelloImpl) Echo(ctx context.Context, req *api.Request) (resp *api.Response, err error) {
	if s.isCrash {
		if s.networkErr { // 如果是网络故障
			<-s.done
			return resp, nil
		}
		// 如果是服务故障
		return &api.Response{Message: fmt.Sprintf("Welcome %s,I am %s", req.Message, s.address)},
			fmt.Errorf("server error")
	}
	resp = &api.Response{Message: fmt.Sprintf("Welcome %s,I am %s", req.Message, s.address)}
	return resp, nil
}
