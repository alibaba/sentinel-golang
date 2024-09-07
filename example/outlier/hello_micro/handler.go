package main

import (
	"context"

	proto "github.com/alibaba/sentinel-golang/pkg/adapters/micro/test"
)

type TestHandler struct{}

func (h *TestHandler) Ping(ctx context.Context, req *proto.Request, rsp *proto.Response) error {
	rsp.Result = "Pong"
	return nil
}
