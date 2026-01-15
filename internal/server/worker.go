package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/worker/pkg/database"
	"github.com/ensoria/worker/pkg/history"
	"github.com/ensoria/worker/pkg/job"
	"github.com/ensoria/worker/pkg/queue"
	"github.com/ensoria/worker/pkg/worker"
	goredis "github.com/redis/go-redis/v9"
)

func NewWorker(lc dikit.LC, cacheClient *goredis.Client, dbClient database.DatabaseClient) worker.Enqueuer {

	qStorage := queue.NewRedisQueue(cacheClient)
	historyRepo := history.NewRepository(dbClient)
	w := worker.New(
		qStorage,
		worker.WithHistory(historyRepo),
	)

	RegisterDefaultJobs(w)

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

// ここでジョブを登録していく
func RegisterDefaultJobs(w *worker.Worker) {
	// SimpleLogJob - シンプルなログ出力
	w.Register("simple_log", SimpleLogHandler, &job.Option{
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		Timeout:    30 * time.Second,
	})

}

// SimpleLogJob - シンプルなログ出力ジョブ
// ペイロードを受け取り、ログに出力するだけのシンプルなジョブ
type SimpleLogPayload struct {
	Message string `json:"message"`
}

func SimpleLogHandler(ctx context.Context, payload []byte) error {
	var p SimpleLogPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	log.Printf("[SimpleLogJob] Processing message: %s", p.Message)

	// シミュレート処理時間
	time.Sleep(500 * time.Millisecond)

	log.Printf("[SimpleLogJob] Completed: %s", p.Message)
	return nil
}
