package bootstrap

import (
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/infra/cache"
	"github.com/ensoria/projecttemplate/internal/infra/db"
	_ "github.com/ensoria/projecttemplate/internal/infra/grpcclt"
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
		cache.NewDefaultRedisClient(envVal),
		db.NewDefaultDatabaseClient(envVal),

		// application
		NewHTTPApp(envVal),
		NewGRPCApp(envVal),
		dikit.InjectHTTPModules(CreateHTTPPipeline),
		dikit.InjectWSModules(CreateWSRouter),

		// worker
		NewWorker,

		// メッセージブローカーのSubscriber接続を提供
		NewSubscriberApp(envVal),
		NewSubscribe,
	})

	dikit.AppendInvocations([]any{
		dikit.RegisterGRPCServices(),
	})

	// TODO: putputFxLogは、環境変数で変えれるようにする
	dikit.ProvideAndRun(dikit.Constructors(), dikit.Invocations(), true)
}
