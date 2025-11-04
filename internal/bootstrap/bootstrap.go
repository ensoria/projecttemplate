package bootstrap

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/ensoria/config/pkg/env"
	"github.com/ensoria/config/pkg/registry"
	_ "github.com/ensoria/projecttemplate/internal/infra/connection/grpcclt"
	_ "github.com/ensoria/projecttemplate/internal/module"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/grpckit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"github.com/ensoria/rest/pkg/mw"
	"github.com/ensoria/rest/pkg/pipeline"
	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/websocket/pkg/wsconfig"
	"github.com/ensoria/websocket/pkg/wsrouter"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// TODO: 別のファイルに分ける
type GlobalError struct {
	Message string `json:"message"`
}

// これを、cmd/main.goで実行する
func Run(envVal *string) {
	registry.InitializeConfiguration(envVal, "./internal", "config")

	dikit.AppendConstructors([]any{
		NewHTTPApp(envVal),
		NewGRPCApp(envVal),
		dikit.AsHTTPPipeline(CreateHTTPPipeline),
		dikit.AsWSRouter(CreateWSRouter),
	})
	// TODO: メッセージブローカーは、httpサーバーとは別にライフサイクル用のcontext.Contextが必要
	// fxのライフサイクルで、メッセージブローカーのcontext.Contextを含める必要がある
	// TODO: putputFxLogは、環境変数で変えれるようにする
	dikit.ProvideAndRun(dikit.Constructors(), dikit.RegisterGRPCServices(), true)
}

// HTTPサーバーの初期化
func NewHTTPApp(envVal *string) func(lc dikit.LC, httpPipeline *pipeline.HTTP, wsRouter *wsrouter.Router) *http.Server {
	return func(lc dikit.LC, httpPipeline *pipeline.HTTP, wsRouter *wsrouter.Router) *http.Server {
		httpPipeline.Register()
		wsRouter.Register()

		// TODO: envValを使うこと
		params, err := registry.ModuleParams("default")
		if err != nil {
			log.Fatalf("default config parameters not found: %s", err)
		}
		// FIXME: 別の場所に移す
		logkit.SetLogLevel(params.Log.Level)

		httpSrv := &http.Server{
			Addr: fmt.Sprintf(":%d", params.Server.Port),
		}

		dikit.RegisterHTTPServerLifecycle(lc, httpSrv)
		return httpSrv
	}
}

func CreateHTTPPipeline(modules []*rest.Module) *pipeline.HTTP {
	// TODO: 別のファイルに分ける
	panicResponse := &rest.Response{
		Code: http.StatusInternalServerError,
		Body: &GlobalError{Message: "internal server error"},
	}

	configParams, err := registry.ModuleParams("default")
	if err != nil {
		log.Fatalf("default config parameters not found: %s", err)
	}
	cors := configParams.Cors

	return &pipeline.HTTP{
		Modules: modules,
		GlobalMiddlewares: []rest.Middleware{
			mw.Logging(logIncomingRequest),
			mw.RecoveryWithLogger(panicResponse, logPanicDetails),
			mw.VerifyBodyParsable,
			mw.NewSimpleCors(cors),
		},
	}
}

func CreateWSRouter(modules []*wsconfig.Module) *wsrouter.Router {
	return &wsrouter.Router{
		Modules: modules,
	}
}

// gRPCサーバーの初期化
func NewGRPCApp(envVal *string) func(lc dikit.LC) *grpc.Server {
	return func(lc dikit.LC) *grpc.Server {
		// ログとpanicリカバリinterceptor付きのgRPCサーバーを作成
		grpcSrv := grpckit.NewGRPCServer(logkit.Logger())

		// reflectionは開発環境でのみ有効にする
		// TODO: config/env にIsLocal()を作って、それを使う
		if envVal != nil && (*envVal == string(env.Local) || *envVal == string(env.Develop)) {
			reflection.Register(grpcSrv)
			logkit.Info("gRPC reflection enabled for development environment", "env", *envVal)
		}

		dikit.RegisterGRPCServerLifecycle(lc, grpcSrv)
		return grpcSrv
	}
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
