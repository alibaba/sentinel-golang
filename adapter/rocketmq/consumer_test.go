package rocketmq

import (
	"context"
	"testing"

	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/stretchr/testify/assert"
)

func TestSentinelConsumerInterceptor(t *testing.T) {
	_, _ = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "group:topic",
			MetricType:      flow.QPS,
			Count:           0,
			ControlBehavior: flow.Reject,
		},
	})
	interceptor := SentinelConsumerInterceptor()

	ctx := primitive.WithConsumerCtx(context.Background(), &primitive.ConsumeMessageContext{
		ConsumerGroup: "group",
		Msgs:          []*primitive.MessageExt{
			{
				Message:                   primitive.Message{
					Topic:         "topic",
					Body:          []byte("hello"),
				},
			},
		},
		MQ:            &primitive.MessageQueue{
			Topic:      "topic",
			BrokerName: "brokerName",
			QueueId:    0,
		},
	})

	err := interceptor(ctx, nil, nil, func(ctx context.Context, req, reply interface{}) error {
		return nil
	})

	assert.NotEqual(t, nil, err)
}
