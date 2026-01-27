package scheduler

import (
	"context"
	"fmt"

	"github.com/ensoria/projecttemplate/internal/app/scheduler/task"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"github.com/ensoria/scheduler/pkg/control"
	"github.com/ensoria/scheduler/pkg/database"
	"github.com/ensoria/scheduler/pkg/distributed"
	"github.com/ensoria/scheduler/pkg/recovery"
	"github.com/ensoria/scheduler/pkg/scheduler"
	sched "github.com/ensoria/scheduler/pkg/scheduler"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

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

func NewSchedulerApp(lc dikit.LC, s *scheduler.Scheduler, tasks []*task.ScheduledTask) error {
	// TODO: httpサーバーの起動も必要そう

	RegisterTasks(s, tasks)

	// REFACTOR: RegisterSchedulerLifeCycle()関数に移動
	lc.Append(dikit.Hook{
		OnStart: func(ctx context.Context) error {
			// TODO: httpサーバーも一緒にlifecycleで管理する?それとも分けるか?
			fmt.Println("Starting scheduler...")
			// OnStartのctxは起動フェーズ用なので、OnStart が完了すると（あるいはタイムアウトすると）キャンセルされる
			// スケジューラーのような長時間実行するサービスには、独立した context.Background() を渡す必要がある
			// これにより、OnStartが完了してもスケジューラーは動き続ける
			if err := s.Start(context.Background()); err != nil {
				return fmt.Errorf("failed to start scheduler: %w", err)
			}
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// TODO: httpサーバーも一緒にlifecycleで管理する?それとも分けるか?

			// シャットダウンcontextは、ctxを使う
			if err := s.Shutdown(ctx); err != nil {
				return fmt.Errorf("scheduler shutdown error: %v", err)
			}
			return nil
		},
	})

	return nil
}

func RegisterTasks(s *sched.Scheduler, tasks []*task.ScheduledTask) error {
	for _, task := range tasks {
		if err := s.SetSchedule(task.Name, task.Cron, task.Task); err != nil {
			return err
		}
	}

	return nil
}

func InjectScheduledTasks(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(``, ``, dikit.GroupTagScheduledTasks),
	)
}
