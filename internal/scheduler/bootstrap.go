package scheduler

import (
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/infra/cache"
	"github.com/ensoria/projecttemplate/internal/infra/db"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"github.com/ensoria/scheduler/pkg/control"
	"github.com/ensoria/scheduler/pkg/database"
	"github.com/ensoria/scheduler/pkg/distributed"
	"github.com/ensoria/scheduler/pkg/recovery"
	"github.com/ensoria/scheduler/pkg/scheduler"
	goredis "github.com/redis/go-redis/v9"
)

func Start(envVal *string) {
	registry.InitializeConfiguration(envVal, "./internal", "config")

	dikit.AppendConstructors([]any{
		// infra
		cache.NewDefaultSchedulerCacheClient(envVal),
		db.NewDefaultSchedulerDBClient(envVal),

		// scheduler
		NewScheduler,
		// TODO: httpサーバーは必要だが、scheduler管理用のエンドポイントのみにする

	})

	dikit.AppendInvocations([]any{
		RegisterSchedulerLifeCycle,
	})

	// TODO: putputFxLogは、環境変数で変えれるようにする
	dikit.ProvideAndRun(dikit.Constructors(), dikit.Invocations(), true)
}

func NewScheduler(
	redisClient *goredis.Client,
	dbClient database.DatabaseClient,
) (*scheduler.Scheduler, error) {
	backend, err := distributed.NewBackend(&distributed.Config{
		BackendType: distributed.BackendTypeRedis,
		Client:      redisClient,
	})
	if err != nil {
		return nil, err
	}

	recoveryRepo := recovery.NewRedisStateRepository(redisClient)
	controlRepo := control.NewRedisStateRepository(redisClient, "")

	s := scheduler.New(backend,
		scheduler.WithLogger(logkit.Logger()),
		scheduler.WithRecovery(recoveryRepo),
		scheduler.WithControl(controlRepo),
		scheduler.WithHistory(dbClient),
	)

	return s, nil

}

func RegisterSchedulerLifeCycle(lc dikit.LC, s *scheduler.Scheduler) error {
	// TODO: httpサーバーの起動も必要そう

	return nil
}
