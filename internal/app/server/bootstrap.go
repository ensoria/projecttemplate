package server

import (
	"github.com/ensoria/config/pkg/registry"
	mbApp "github.com/ensoria/projecttemplate/internal/app/mb"
	workerApp "github.com/ensoria/projecttemplate/internal/app/worker"
	"github.com/ensoria/projecttemplate/internal/infra/cache"
	"github.com/ensoria/projecttemplate/internal/infra/db"
	_ "github.com/ensoria/projecttemplate/internal/infra/grpcclt"
	"github.com/ensoria/projecttemplate/internal/infra/mb"
	_ "github.com/ensoria/projecttemplate/internal/infra/mb"
	_ "github.com/ensoria/projecttemplate/internal/module"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
)

// TODO: 別のファイルに分ける
type GlobalError struct {
	Message string `json:"message"`
}

func Run(envVal *string) {
	registry.InitializeConfiguration(envVal, "./internal", "config")

	dikit.AppendConstructors([]any{
		// infra
		cache.NewDefaultWorkerCacheClient(envVal),
		db.NewDefaultWorkerDBClient(envVal),
		mb.NewSubscriberConnection(envVal),
		mb.NewPublisherConnection(envVal),

		// controllers
		dikit.InjectHTTPModules(CreateHTTPPipeline),
		dikit.InjectWSModules(CreateWSRouter),
		dikit.InjectGRPCServices(CreateGRPCServices),
		mbApp.NewSubscribe,
		mbApp.NewPublish,

		// worker
		dikit.InjectWorkerJobs(workerApp.NewWorker),
	})

	dikit.AppendInvocations([]any{
		// application invocations
		NewHTTPApp(envVal),
		NewGRPCApp(envVal),
	})

	// TODO: 最後の引数の、putputFxLogは、環境変数で変えれるようにする
	dikit.ProvideAndRun(dikit.Constructors(), dikit.Invocations(), true)
}
