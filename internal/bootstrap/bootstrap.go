package bootstrap

import (
	"fmt"
	"log/slog"

	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/infra/cache"
	"github.com/ensoria/projecttemplate/internal/infra/db"
	_ "github.com/ensoria/projecttemplate/internal/infra/grpcclt"
	_ "github.com/ensoria/projecttemplate/internal/infra/mb"
	_ "github.com/ensoria/projecttemplate/internal/module"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"github.com/ensoria/rest/pkg/rest"
)

// TODO: 別のファイルに分ける
type GlobalError struct {
	Message string `json:"message"`
}

// これを、cmd/main.goで実行する
func Run(envVal *string) {
	registry.InitializeConfiguration(envVal, "./internal", "config")

	dikit.AppendConstructors([]any{
		// infra
		cache.NewDefaultRedisClient,
		db.NewDefaultDatabaseClient,

		// application
		NewHTTPApp(envVal),
		NewGRPCApp(envVal),
		dikit.InjectHTTPModules(CreateHTTPPipeline),
		dikit.InjectWSModules(CreateWSRouter),

		// worker
		NewQueueStorage,
		NewHistoryRepository,
		NewWorker(envVal),

		// メッセージブローカーのSubscriber接続を提供
		NewSubscriberApp(envVal),
		NewSubscribe,
	})

	// TODO: constructorとinvocationの使い分けがいまいち分からないので調べる
	dikit.AppendInvocations([]any{
		dikit.RegisterGRPCServices(),
	})

	// TODO: putputFxLogは、環境変数で変えれるようにする
	dikit.ProvideAndRun(dikit.Constructors(), dikit.Invocations(), true)
}

func logIncomingRequest(req *rest.Request, res *rest.Response) {
	logkit.Info("HTTP Request",
		slog.String("method", req.Method()),
		slog.String("path", req.Path()),
		slog.Int("status_code", res.Code),
		slog.String("remote_addr", req.RemoteAddr()),
		slog.String("user_agent", req.UserAgent()),
		slog.String("type", "access_log"),
	)
}

func logPanicDetails(r *rest.Request, panicValue interface{}, stackTrace []byte) {
	logkit.Error("Panic Recovered",
		slog.String("method", r.Method()),
		slog.String("url", r.URLStr()),
		slog.String("remote_addr", r.RemoteAddr()),
		slog.String("user_agent", r.UserAgent()),
		slog.Any("panic_value", panicValue),
		slog.String("panic_type", fmt.Sprintf("%T", panicValue)),
		slog.String("stack_trace", string(stackTrace)),
		slog.String("type", "panic_log"),
	)
}
