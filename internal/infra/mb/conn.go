package mb

import (
	"context"
	"fmt"
	"log/slog"

	enmb "github.com/ensoria/mb/pkg/mb"
	"github.com/ensoria/mb/pkg/mq"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
)

// message brokerに関する接続

type SubscriberPanicHandler struct{}

func (h *SubscriberPanicHandler) OnPanic(panicValue interface{}, stackTrace []byte, metadata enmb.PanicMetadata) {
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

func NewSubscriberConnection(envVal *string) func(lc dikit.LC) (enmb.Subscriber, error) {
	return func(lc dikit.LC) (enmb.Subscriber, error) {
		// TODO: envValを使って、その環境の値をconfigから取得するようにする
		// configにはメッセージブローカーの実装がないので、configで実装してから変更
		config := &enmb.Config{
			Type: enmb.TypeRabbitMQ,
			URL:  "amqp://localhost:5672/",
			Credentials: &enmb.Credentials{
				Username: "myuser",
				Password: "mypassword",
			},
		}

		subConn, err := mq.NewSubscriber(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create subscriber: %w", err)
		}

		subConn.SetOptions(
			enmb.WithLogger(logkit.Logger()),
			enmb.WithPanicHandler(&SubscriberPanicHandler{}),
		)

		onStop := func(ctx context.Context) error {
			slog.Info("Shutting down MB subscriber")
			return subConn.Close()
		}
		dikit.RegisterOnStopLifecycle(lc, onStop)

		return subConn, nil
	}
}

func NewPublisherConnection(envVal *string) func(lc dikit.LC) (enmb.Publisher, error) {
	return func(lc dikit.LC) (enmb.Publisher, error) {
		// configから取得する
		config := &enmb.Config{
			Type: enmb.TypeRabbitMQ,
			URL:  "amqp://localhost:5672/",
			Credentials: &enmb.Credentials{
				Username: "myuser",
				Password: "mypassword",
			},
		}

		pubConn, err := mq.NewPublisher(config)
		if err != nil {
			return nil, fmt.Errorf("failed to create publisher: %w", err)
		}

		pubConn.SetOptions(enmb.WithPublishLogger(logkit.Logger()))

		onStop := func(ctx context.Context) error {
			logkit.Info("Shutting down MB publisher")
			return pubConn.Close()
		}
		dikit.RegisterOnStopLifecycle(lc, onStop)

		return pubConn, nil
	}
}
