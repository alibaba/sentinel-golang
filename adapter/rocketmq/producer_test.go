package rocketmq

import (
	"context"
	"testing"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	_ = sentinel.InitDefault()
	m.Run()
}

func TestSentinelProviderInterceptor(t *testing.T) {
	_, _ = flow.LoadRules([]*flow.FlowRule{
		{
			Resource:        "group:topic",
			MetricType:      flow.QPS,
			Count:           0,
			ControlBehavior: flow.Reject,
		},
	})
	interceptor := SentinelProviderInterceptor()

	ctx := primitive.WithProducerCtx(context.Background(), &primitive.ProducerCtx{
		ProducerGroup:     "group",
		Message:           primitive.Message{
			Topic:         "topic",
			Body:          []byte("hello"),
			Flag:          0,
			TransactionId: "",
			Batch:         false,
			Queue:         nil,
		},
		MQ:                primitive.MessageQueue{
			Topic:      "topic",
			BrokerName: "brokerName",
			QueueId:    10000,
		},
	})

	err := interceptor(ctx, nil, nil, func(ctx context.Context, req, reply interface{}) error {
		return nil
	})

	assert.NotEqual(t, nil, err)
}