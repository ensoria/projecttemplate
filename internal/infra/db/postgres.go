package db

import (
	"context"

	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/worker/pkg/database"
)

func NewDefaultDatabaseClient(envVal *string) func(lc dikit.LC) (database.DatabaseClient, error) {
	return func(lc dikit.LC) (database.DatabaseClient, error) {
		// TODO: envValとconfigパッケージを使って設定を取得するようにする
		dbConfig := &database.DatabaseConfig{
			Type:     database.DBTypePostgreSQL,
			Host:     "localhost",
			Port:     5432,
			User:     "ensoria",
			Password: "ensoria",
			Database: "ensoria",
			TLSMode:  "disable",
		}

		client, err := database.NewDatabaseClient(dbConfig)
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
