package flow

import (
	"context"
	"github.com/sentinel-group/sentinel-golang/core/slots/base"
)

type TrafficShapingController interface {
	CanPass(ctx context.Context, node *base.DefaultNode, acquire uint32) bool
}

type WarmUpController struct {
}

func (wpc WarmUpController) CanPass(ctx context.Context, node *base.DefaultNode, acquire uint32) bool {
	return true
}

type RateLimiterController struct {
}

func (wpc RateLimiterController) CanPass(ctx context.Context, node *base.DefaultNode, acquire uint32) bool {
	return true
}

type WarmUpRateLimiterController struct {
}

func (wpc WarmUpRateLimiterController) CanPass(ctx context.Context, node *base.DefaultNode, acquire uint32) bool {
	return true
}

type DefaultController struct {
}

func (wpc DefaultController) CanPass(ctx context.Context, node *base.DefaultNode, acquire uint32) bool {
	return true
}
