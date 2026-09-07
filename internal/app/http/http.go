package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ensoria/config/pkg/appconfig"
	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/ensoria-template/internal/app/http/dto"
	"github.com/ensoria/ensoria-template/internal/middleware"
	"github.com/ensoria/ensoria-template/internal/plamo/apidoc"
	"github.com/ensoria/ensoria-template/internal/plamo/authkit"
	"github.com/ensoria/ensoria-template/internal/plamo/dikit"
	"github.com/ensoria/ensoria-template/internal/plamo/restkit"
	"github.com/ensoria/loggear/pkg/loggear"
	"github.com/ensoria/rest/pkg/mw"
	"github.com/ensoria/rest/pkg/pipeline"
	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/websocket/pkg/wsconn"
	"github.com/ensoria/websocket/pkg/wsrouter"
	"go.uber.org/fx"
)

// HTTPサーバーの初期化
//
// A failure is returned rather than written with log.Fatal: fx reports it
// through the one path every other startup failure takes, and the lifecycle
// still gets to stop what had already started — os.Exit would leave the
// connections opened before this point unclosed.
func NewHTTPApp(envVal *string) func(lc dikit.LC, shutdowner dikit.Shutdowner, httpPipeline *pipeline.HTTP, wsRouter *wsrouter.Router) (*http.Server, error) {
	return func(lc dikit.LC, shutdowner dikit.Shutdowner, httpPipeline *pipeline.HTTP, wsRouter *wsrouter.Router) (*http.Server, error) {
		// TODO: envValを使うこと
		params, err := registry.ModuleParams("default")
		if err != nil {
			return nil, fmt.Errorf("default config parameters not found: %w", err)
		}

		// HTTPパイプラインとWebSocketルータを同一のmuxに登録する。
		// グローバルなhttp.DefaultServeMuxを使わないことで、ハンドラを分離でき、
		// テストの並列化や複数サーバの起動が可能になる。
		mux := http.NewServeMux()
		httpPipeline.Register(mux)
		wsRouter.Register(mux)

		httpSrv := &http.Server{
			Addr:    fmt.Sprintf(":%d", params.Server.Port),
			Handler: mux,
			// Layer 1: コネクションレベルのタイムアウト（configから取得）
			ReadHeaderTimeout: params.Server.ReadHeaderTimeout,
			ReadTimeout:       params.Server.ReadTimeout,
			// WriteTimeoutはレスポンス書き込み全体の絶対deadlineであり、SSE・WebSocket・
			// 大きなファイルのような長時間接続を切断する。そのため既定では0(無効)。
			// リクエスト単位のタイムアウトはpipeline側(Layer 2)で制御する。
			WriteTimeout: params.Server.WriteTimeout,
			IdleTimeout:  params.Server.IdleTimeout,
		}

		RegisterHTTPServerLifecycle(lc, shutdowner, httpSrv, wsRouter)
		return httpSrv, nil
	}
}

func CreateHTTPPipeline(envVal *string) func(modules []*rest.Module, verifier authkit.Verifier, origins *middleware.Origins) (*pipeline.HTTP, error) {
	return func(modules []*rest.Module, verifier authkit.Verifier, origins *middleware.Origins) (*pipeline.HTTP, error) {
		return createHTTPPipeline(*envVal, modules, verifier, origins)
	}
}

func createHTTPPipeline(
	envVal string, modules []*rest.Module, verifier authkit.Verifier, origins *middleware.Origins,
) (*pipeline.HTTP, error) {
	// TODO: 別のファイルに分ける
	panicResponse := &rest.Response{
		Code: http.StatusInternalServerError,
		Body: &dto.Error{Message: "internal server error"},
	}

	configParams, err := registry.ModuleParams("default")
	if err != nil {
		return nil, fmt.Errorf("default config parameters not found: %w", err)
	}

	// 宣言と設定の食い違いを起動時に潰す。放っておくと全リクエストが拒否されるだけで、
	// 原因が見えない。
	//
	// The session checks run first. They are the specific case of the same
	// mistake, and their messages name the wiring rather than the symptom.
	if err := checkSessionConfiguration(envVal, configParams, modules); err != nil {
		return nil, err
	}
	if err := checkAuthConfiguration(modules, verifier); err != nil {
		return nil, err
	}

	// CORS, this check and the WebSocket upgrade guard are all handed the same
	// resolved origins (see ws.NewTrustedOrigins) rather than each reading
	// CORS_ALLOW_ORIGIN for themselves.
	crossOrigin, err := middleware.NewCrossOriginProtection(origins)
	if err != nil {
		return nil, err
	}

	// Layer 2: リクエスト単位（ハンドラ実行）のタイムアウト超過時に返すレスポンス
	timeoutResponse := &rest.Response{
		Code: http.StatusServiceUnavailable,
		Body: &dto.Error{Message: "request timeout"},
	}

	return &pipeline.HTTP{
		Modules: modules,
		GlobalMiddlewares: globalMiddlewares(&globalMiddlewareDeps{
			cors:          configParams.CORS,
			crossOrigin:   crossOrigin,
			verifier:      verifier,
			panicResponse: panicResponse,
		}),
		// Layer 2: コントローラ/ミドルウェアチェーンの実行（=レスポンスの計算）の上限時間。
		// 0で無効。ストリーミング/ファイル/WebSocketは対象外。
		Timeout:         configParams.Server.HandlerTimeout,
		TimeoutResponse: timeoutResponse,
	}, nil
}

// HTTP/WebSocket controllers lifecycle registration
func RegisterHTTPServerLifecycle(lc dikit.LC, shutdowner dikit.Shutdowner, srv *http.Server, wsRouter *wsrouter.Router) {
	lc.Append(dikit.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				loggear.Info("HTTP server starting", "addr", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					loggear.Error("HTTP server stopped unexpectedly", "error", err)
					// The exit code has to say the process failed. Shutting down
					// without one ends the process with 0, and a supervisor that
					// restarts on failure (systemd Restart=on-failure, a
					// Kubernetes restartPolicy) reads that as a clean stop and
					// leaves the application down.
					_ = shutdowner.Shutdown(dikit.ExitCode(1))
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// WebSocketを先に閉じる。Upgradeでハイジャックされた接続は
			// http.Serverの管理外なので、srv.Shutdownは待っても閉じてもくれない。
			// wsRouter.Shutdownが各サーバのbaseCtxをキャンセルして全接続の
			// connCtxに伝播させ（進行中の処理を中断可能にし）、close frameを
			// 送って読み取りループを解く。各接続のOnCloseは接続ctxとは切り離された
			// ctx（OnCloseTimeout）で後始末を完走できる。
			closed := wsRouter.Shutdown(wsconn.CloseGoingAway, "server shutting down")
			loggear.Info("Closed WebSocket connections", "count", closed)

			loggear.Info("Shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}

// globalMiddlewareDeps carries what the chain is built from. It is a struct
// rather than four parameters so that an entry below can ignore what it does not
// need without every entry restating the whole list.
type globalMiddlewareDeps struct {
	cors          *appconfig.CORS
	crossOrigin   middleware.CrossOriginChecker
	verifier      authkit.Verifier
	panicResponse *rest.Response
}

// globalMiddleware is one link of the chain: the name it is known by outside
// this package, and how to build it once its dependencies exist.
//
// Build is a function rather than a ready middleware because the names are
// needed where the dependencies are not — the document generator resolves no
// configuration, opens no store and has no verifier, yet has to say what runs.
type globalMiddleware struct {
	Name  string
	Build func(deps *globalMiddlewareDeps) rest.Middleware
}

// globalMiddlewareChain is the single statement of what runs on every request,
// in order. The pipeline and the generated documentation are both projections
// of it — globalMiddlewares takes the Build side, GlobalMiddlewareNames the Name
// side — so a middleware added here reaches both, and one cannot describe a
// pipeline the other does not run.
//
// It was two lists until 2026-09-07, and they had already drifted: the names
// stayed at four entries after CSRF and Auth joined the chain, so every
// generated document described a pipeline the application had stopped running
// two phases earlier. Nothing could have caught it, because nothing in the type
// system relates a []rest.Middleware to a []string.
//
// # Order
//
// The chain runs outside-in, so authentication sits last: logging, panic
// recovery and CORS still apply to a request that is refused, and a CORS
// preflight (which carries no credential) is answered before authentication is
// considered.
//
// The cross-origin check sits between CORS and authentication. Before
// authentication, so a forged request is refused without the session store being
// asked about the cookie it carried; after CORS, so a preflight is still
// answered by the one middleware that knows how to answer it.
//
// ⚠ Only one of those two refuses anything. CORS tells the browser what it may
// read and refuses nothing; the cross-origin check refuses state-changing
// requests from origins this deployment does not claim. See middleware.CORS for
// why the split is that way round.
var globalMiddlewareChain = []globalMiddleware{
	{
		Name:  apidoc.MiddlewareLogging,
		Build: func(*globalMiddlewareDeps) rest.Middleware { return mw.Logging(logIncomingRequest) },
	},
	{
		Name: apidoc.MiddlewareRecovery,
		Build: func(d *globalMiddlewareDeps) rest.Middleware {
			return mw.RecoveryWithLogger(d.panicResponse, logPanicDetails)
		},
	},
	{
		Name:  apidoc.MiddlewareVerifyBodyParsable,
		Build: func(*globalMiddlewareDeps) rest.Middleware { return mw.VerifyBodyParsable },
	},
	{
		Name:  apidoc.MiddlewareCORS,
		Build: func(d *globalMiddlewareDeps) rest.Middleware { return middleware.CORS(d.cors) },
	},
	{
		Name:  apidoc.MiddlewareCSRF,
		Build: func(d *globalMiddlewareDeps) rest.Middleware { return middleware.CSRF(d.crossOrigin) },
	},
	{
		Name:  apidoc.MiddlewareAuth,
		Build: func(d *globalMiddlewareDeps) rest.Middleware { return middleware.Auth(d.verifier) },
	},
}

// globalMiddlewares builds the chain every request passes through, in the order
// globalMiddlewareChain declares.
func globalMiddlewares(deps *globalMiddlewareDeps) []rest.Middleware {
	chain := make([]rest.Middleware, 0, len(globalMiddlewareChain))
	for _, m := range globalMiddlewareChain {
		chain = append(chain, m.Build(deps))
	}
	return chain
}

// GlobalMiddlewareNames names, in order, what the chain installs.
//
// The generated documentation reads it, and one entry in particular is acted on:
// a document describing a cookie-borne credential explains the cross-origin
// guard only when this list names it. A name missing here would understate what
// the application does, and a name outliving its middleware would promise a
// protection that is gone — which is why the names are derived from the chain
// rather than kept beside it.
//
// A fresh slice is returned because the caller hands it to a document model that
// outlives the call; sharing the backing array would let a renderer sort or
// truncate the declaration itself.
func GlobalMiddlewareNames() []string {
	names := make([]string, 0, len(globalMiddlewareChain))
	for _, m := range globalMiddlewareChain {
		names = append(names, m.Name)
	}
	return names
}

// InjectHTTPModules tags the first parameter as the HTTP module group. The
// remaining parameters (the credential verifier) are resolved by type.
func InjectHTTPModules(f any) any {
	return fx.Annotate(f, fx.ParamTags(dikit.GroupTagHttpModules, ``))
}

func logIncomingRequest(req *rest.Request, res *rest.Response) {
	loggear.Info("HTTP Request",
		slog.String("method", req.Method()),
		slog.String("path", req.Path()),
		slog.Int("status_code", res.Code),
		slog.String("remote_addr", req.RemoteAddr()),
		slog.String("user_agent", req.UserAgent()),
		slog.String("type", "access_log"),
	)
}

// panicViolationGroup is the key a contract violation's fields are nested under
// in a panic record.
//
// They are grouped rather than written flat beside the record's own fields
// because a violation names the endpoint as well: expanded flat, "method" would
// appear in one JSON object twice and neither occurrence could be relied on.
// Grouping also leaves "type" saying panic_log and nothing else, which is what a
// panic alert matches on.
const panicViolationGroup = "contract_violation"

// logPanicDetails records a panic the recovery middleware caught. It is the
// only account of a request that failed for a reason nobody anticipated, so it
// carries what identifies the defect — the value, its type, the stack — and the
// least that identifies where it came from: the endpoint and the caller's
// address, which is what makes it possible to see what else that caller sent.
//
// It deliberately leaves out the client's user agent. A panic is a defect in
// this server, and which browser asked does not help say which defect it is;
// the access log written for the same request carries it for the cases where
// the client population is the question.
func logPanicDetails(r *rest.Request, panicValue interface{}, stackTrace []byte) {
	args := []any{
		slog.String("method", r.Method()),
		slog.String("url", r.URLStr()),
		slog.String("remote_addr", r.RemoteAddr()),
		slog.Any("panic_value", panicValue),
		slog.String("panic_type", fmt.Sprintf("%T", panicValue)),
		slog.String("stack_trace", string(stackTrace)),
		slog.String("type", "panic_log"),
	}
	// A panic carrying a contract violation can say which promise the
	// implementation broke, which a stack trace of generic frames cannot. The
	// assertion is on the interface rather than on any one kind of violation, so
	// a kind added later is expanded here without this file being touched.
	if violation, ok := panicValue.(restkit.ContractViolation); ok {
		args = append(args, slog.Group(panicViolationGroup, restkit.LogArgs(violation)...))
	}
	loggear.Error("Panic Recovered", args...)
}
