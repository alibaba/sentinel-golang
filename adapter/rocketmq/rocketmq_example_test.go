package rocketmq

import (
	"context"
	"fmt"

	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
)

func Example_Consumer() {
	var c, _ = consumer.NewPushConsumer(
		consumer.WithInterceptor( SentinelConsumerInterceptor()),
		consumer.WithNameServer(primitive.NamesrvAddr{ "127.0.0.1:9876"}),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	)

	err := c.Subscribe( "testTopic",
		consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			fmt.Println(msgs)
			for _, msg := range msgs {
				msg.GetTags()
			}
			return consumer.ConsumeSuccess, nil
		},
	)

	if err != nil {
		// todo something
	}
}

func Example_Provider() {
	var c, _ = consumer.NewPushConsumer(
		consumer.WithInterceptor( SentinelConsumerInterceptor()),
		consumer.WithNameServer(primitive.NamesrvAddr{ "127.0.0.1:9876"}),
		consumer.WithConsumerModel(consumer.Clustering),
		consumer.WithConsumeFromWhere(consumer.ConsumeFromFirstOffset),
	)

	err := c.Subscribe( "testTopic",
		consumer.MessageSelector{},
		func(ctx context.Context, msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
			fmt.Println(msgs)
			return consumer.ConsumeSuccess, nil
		},
	)

	if err != nil {
		// todo something
	}
}
