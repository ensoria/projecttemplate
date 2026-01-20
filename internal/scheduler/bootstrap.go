package scheduler

import (
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/infra/cache"
	"github.com/ensoria/projecttemplate/internal/infra/db"
	"github.com/ensoria/projecttemplate/internal/infra/mb"
	"github.com/ensoria/projecttemplate/internal/server"

	_ "github.com/ensoria/projecttemplate/internal/module" // TODO:
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
)

func Start(envVal *string) {
	registry.InitializeConfiguration(envVal, "./internal", "config")

	dikit.AppendConstructors([]any{
		// infra
		// workerとinjectするインスタンスを分けるため、タグ名を付ける
		dikit.ProvideNamed(cache.NewDefaultSchedulerCacheClient(envVal), "schedulerCache"),
		db.NewDefaultSchedulerDBClient(envVal),

		// TODO: 無くてもいいようにする?
		dikit.ProvideNamed(cache.NewDefaultWorkerCacheClient(envVal), "workerCache"),
		db.NewDefaultWorkerDBClient(envVal),
		mb.NewSubscriberConnection(envVal),
		mb.NewPublisherConnection(envVal),
		server.NewSubscribe,
		server.NewPublish,
		dikit.InjectWithTags(server.NewWorker, ``, `name:"workerCache"`, ``),

		// scheduler
		// タグ名の付いたキャッシュクライアントを注入
		dikit.InjectWithTags(NewScheduler, `name:"schedulerCache"`, ``),
		// TODO: httpサーバーは必要だが、scheduler管理用のエンドポイントのみにする

	})

	dikit.AppendInvocations([]any{
		dikit.InjectScheduledTasks(NewSchedulerApp),
	})

	// TODO: putputFxLogは、環境変数で変えれるようにする
	dikit.ProvideAndRun(dikit.Constructors(), dikit.Invocations(), true)
}
