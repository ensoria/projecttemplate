package cache

import (
	"context"

	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	goredis "github.com/redis/go-redis/v9"
)

func NewDefaultWorkerCacheClient(envVal *string) func(lc dikit.LC) *goredis.Client {
	return func(lc dikit.LC) *goredis.Client {
		// TODO: envValとconfigパッケージを使って設定を取得するようにする
		// params := registry.ModuleParams("default")
		client := goredis.NewClient(&goredis.Options{
			Addr: "localhost:6379",
			DB:   0,
		})

		lc.Append(dikit.Hook{
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

}

func NewDefaultSchedulerCacheClient(envVal *string) func(lc dikit.LC) *goredis.Client {
	return func(lc dikit.LC) *goredis.Client {
		// TODO: envValとconfigパッケージを使って設定を取得するようにする
		// params := registry.ModuleParams("default")
		client := goredis.NewClient(&goredis.Options{
			Addr: "localhost:6379",
			DB:   1,
		})

		lc.Append(dikit.Hook{
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

}
