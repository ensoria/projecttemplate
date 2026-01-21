package mb

import (
	"context"

	"github.com/ensoria/mb/pkg/mb"
)

func NewSubscribe(subConn mb.Subscriber) mb.StartSubscription {
	return func(target string, handler mb.SubscribeHandler, opts ...mb.SubscribeOption) error {
		// SubscribeHandlerのOnReceiveメソッドをMessageHandlerに変換
		messageHandler := func(data []byte, metadata map[string]string) error {
			return handler.OnReceive(data, metadata)
		}
		// fxのライフサイクル内で実行されるため、context.Backgroundを使用
		return subConn.Subscribe(context.Background(), target, messageHandler, opts...)
	}
}

func NewPublish(pubConn mb.Publisher) mb.Publish {
	return func(target string, data []byte, metadata map[string]string, opts ...mb.PublishOption) error {
		// fxのライフサイクル内で実行されるため、context.Backgroundを使用
		return pubConn.Publish(context.Background(), target, data, metadata, opts...)
	}
}
