package worker

import (
	"context"
	"log/slog"

	appJob "github.com/ensoria/projecttemplate/internal/app/worker/job"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/worker/pkg/database"
	"github.com/ensoria/worker/pkg/history"
	"github.com/ensoria/worker/pkg/queue"
	"github.com/ensoria/worker/pkg/worker"
	goredis "github.com/redis/go-redis/v9"
	"go.uber.org/fx"
)

func NewWorker(
	lc dikit.LC,
	cacheClient *goredis.Client,
	dbClient database.DatabaseClient,
	jobs []*appJob.JobHandler,
) worker.Enqueuer {

	qStorage := queue.NewRedisQueue(cacheClient)
	historyRepo := history.NewRepository(dbClient)
	w := worker.New(
		qStorage,
		worker.WithHistory(historyRepo),
	)

	for _, j := range jobs {
		w.Register(j.Name, j.Handler, j.Options)
	}

	// ワーカーのContext
	var workerCtx context.Context
	var workerCancel context.CancelFunc

	lc.Append(dikit.Hook{
		OnStart: func(ctx context.Context) error {
			workerCtx, workerCancel = context.WithCancel(context.Background())
			go func() {
				slog.Info("Starting worker...")
				w.Start(workerCtx)
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Stopping worker...")
			workerCancel()
			return w.Shutdown(ctx)
		},
	})

	return w

}

func InjectWorkerJobs(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(``, ``, ``, dikit.GroupTagWorkerJobs),
	)
}
