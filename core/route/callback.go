package route

import (
	"context"
	"errors"
)

type InitFunc func() error
type GetTrafficTagFunc func(ctx context.Context) string
type GetPodTagFunc func(ctx context.Context) string
type SetTrafficTagFunc func(ctx context.Context, trafficTag string) (context.Context, error)

type Callback struct {
	InitFunc      InitFunc
	GetTrafficTag GetTrafficTagFunc
	GetPodTag     GetPodTagFunc
	SetTrafficTag SetTrafficTagFunc
}

var callBack *Callback

func NewCallbackFunc(initFunc InitFunc, getTrafficTag GetTrafficTagFunc, getPodTag GetPodTagFunc, setTrafficTag SetTrafficTagFunc) error {
	if getTrafficTag == nil || getPodTag == nil || setTrafficTag == nil {
		return errors.New("callback func can not be nil")
	}

	callBack = &Callback{
		InitFunc:      initFunc,
		GetTrafficTag: getTrafficTag,
		GetPodTag:     getPodTag,
		SetTrafficTag: setTrafficTag,
	}

	if initFunc != nil {
		err := initFunc()
		if err != nil {
			return err
		}
	}
	return nil
}

func getTrafficTag(ctx context.Context) string {
	if callBack == nil || callBack.GetTrafficTag == nil {
		return ""
	}
	return callBack.GetTrafficTag(ctx)
}

func getPodTag(ctx context.Context) string {
	if callBack == nil || callBack.GetTrafficTag == nil {
		return ""
	}
	return callBack.GetPodTag(ctx)
}

func setTrafficTag(ctx context.Context, trafficTag string) (context.Context, error) {
	if callBack == nil || callBack.SetTrafficTag == nil {
		return ctx, errors.New("set traffic tag callback func is nil")
	}
	return callBack.SetTrafficTag(ctx, trafficTag)
}
