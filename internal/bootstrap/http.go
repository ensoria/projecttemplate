package bootstrap

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ensoria/config/pkg/registry"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"github.com/ensoria/rest/pkg/mw"
	"github.com/ensoria/rest/pkg/pipeline"
	"github.com/ensoria/rest/pkg/rest"
	"github.com/ensoria/websocket/pkg/wsrouter"
)

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
