package cache

import (
	"context"

	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewDefaultRedisClient(lc dikit.LC) *goredis.Client {
	// TODO: configパッケージを使って設定を取得するようにする
	client := goredis.NewClient(&goredis.Options{
		Addr: "localhost:6379",
		DB:   0,
	})

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := client.Ping(ctx).Err(); err != nil {
				return err
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return client.Close()
		},
	})

	return client
}
