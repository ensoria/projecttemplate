package queue

import (
	"context"
	"fmt"

	"github.com/ensoria/mb/pkg/mb"
	"github.com/ensoria/mb/pkg/mq"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
)

// message brokerに関する接続

func NewPubConnection(lc dikit.LC) (mb.Publisher, error) {
	// configから取得する
	config := &mb.Config{
		Type: mb.TypeRabbitMQ,
		URL:  "amqp://localhost:5672/",
		Credentials: &mb.Credentials{
			Username: "myuser",
			Password: "mypassword",
		},
	}

	pubConn, err := mq.NewPublisher(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create publisher: %w", err)
	}

	pubConn.SetOptions(mb.WithPublishLogger(logkit.Logger()))

	onStop := func(ctx context.Context) error {
		logkit.Info("Shutting down MB publisher")
		return pubConn.Close()
	}
	dikit.RegisterOnStopLifecycle(lc, onStop)

	return pubConn, nil
}

func NewPublish(pubConn mb.Publisher) mb.Publish {
	return func(target string, data []byte, metadata map[string]string, opts ...mb.PublishOption) error {
		// fxのライフサイクル内で実行されるため、context.Backgroundを使用
		return pubConn.Publish(context.Background(), target, data, metadata, opts...)
	}
}

func init() {
	dikit.AppendConstructors([]any{
		NewPubConnection,
		NewPublish,
	})
}
