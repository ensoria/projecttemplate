package server

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/ensoria/mb/pkg/mb"
	"github.com/ensoria/mb/pkg/mq"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
)

func NewSubscriberApp(envVal *string) func(lc dikit.LC) (mb.Subscriber, error) {
	return func(lc dikit.LC) (mb.Subscriber, error) {
		// TODO: envValを使って、その環境の値をconfigから取得するようにする
		// configにはメッセージブローカーの実装がないので、configで実装してから変更
		config := &mb.Config{
			Type: mb.TypeRabbitMQ,
			URL:  "amqp://localhost:5672/",
			Credentials: &mb.Credentials{
				Username: "myuser",
				Password: "mypassword",
			},
		}

		subConn, err := mq.NewSubscriber(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscriber: %w", err)
		}

		subConn.SetOptions(
			mb.WithLogger(logkit.Logger()),
			mb.WithPanicHandler(&SubscriberPanicHandler{}),
		)

		onStop := func(ctx context.Context) error {
			slog.Info("Shutting down MB subscriber")
			return subConn.Close()
		}
		dikit.RegisterOnStopLifecycle(lc, onStop)

		return subConn, nil
	}
}

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

type SubscriberPanicHandler struct{}

func (h *SubscriberPanicHandler) OnPanic(panicValue interface{}, stackTrace []byte, metadata mb.PanicMetadata) {
	logkit.Error("Panic Recovered in Subscriber",
		"target", metadata.Target,
		"metadata", metadata.Metadata,
		"data", metadata.Data,
		"panic_value", panicValue,
		"panic_type", fmt.Sprintf("%T", panicValue),
		"stack_trace", string(stackTrace),
		"type", "subscriber_panic_log",
	)
}
