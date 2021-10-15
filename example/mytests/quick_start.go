package main

import (
	"fmt"
	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
)

func main() {
	if err := sentinel.InitDefault(); err != nil {
		// 初始化失败
		panic(err.Error())
	}

	// 资源名
	resource := "test-resource"

	// 加载流控规则，写死
	_, err := flow.LoadRules([]*flow.Rule{
		{
			Resource: resource,
			// Threshold + StatIntervalInMs 可组合出多长时间限制通过多少请求，这里相当于限制为 10 qps
			Threshold:        10,
			StatIntervalInMs: 1000,
			// 暂时不用关注这些参数
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
		},
	})

	if err != nil {
		panic(err.Error())
	}

	// 修改这个看看效果吧
	currency := 100

	for i := 0; i < currency; i++ {
		go func() {
			e, b := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound))
			if b != nil {
				// 被流控
				fmt.Printf("blocked %s \n", b.BlockMsg())
			} else {
				// 通过
				fmt.Println("pass...")
				// 通过后必须调用Exit
				e.Exit()
			}
		}()
	}
}
