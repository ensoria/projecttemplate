package db

import (
	"context"

	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	schedulerDB "github.com/ensoria/scheduler/pkg/database"
	workerDB "github.com/ensoria/worker/pkg/database"
)

func NewDefaultWorkerDBClient(envVal *string) func(lc dikit.LC) (workerDB.DatabaseClient, error) {
	return func(lc dikit.LC) (workerDB.DatabaseClient, error) {
		// TODO: envValとconfigパッケージを使って設定を取得するようにする
		// params := registry.ModuleParams("default")
		dbConfig := &workerDB.DatabaseConfig{
			Type:     workerDB.DBTypePostgreSQL,
			Host:     "localhost",
			Port:     5432,
			User:     "ensoria",
			Password: "ensoria",
			Database: "ensoria",
			TLSMode:  "disable",
		}

		client, err := workerDB.NewDatabaseClient(dbConfig)
		if err != nil {
			return nil, err
		}

		lc.Append(dikit.Hook{
			OnStart: func(ctx context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return client.Close()
			},
		})

		return client, nil
	}
}

func NewDefaultSchedulerDBClient(envVal *string) func(lc dikit.LC) (schedulerDB.DatabaseClient, error) {
	return func(lc dikit.LC) (schedulerDB.DatabaseClient, error) {
		// TODO: envValとconfigパッケージを使って設定を取得するようにする
		// params := registry.ModuleParams("default")
		cfg := &schedulerDB.DatabaseConfig{
			Type:     schedulerDB.DBTypePostgreSQL,
			Host:     "localhost",
			Port:     5432,
			User:     "ensoria",
			Password: "ensoria",
			Database: "ensoria",
			TLSMode:  "disable",
		}

		client, err := schedulerDB.NewDatabaseClient(cfg)
		if err != nil {
			return nil, err
		}

		lc.Append(dikit.Hook{
			OnStart: func(ctx context.Context) error {
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return client.Close()
			},
		})

		return client, nil
	}

}
