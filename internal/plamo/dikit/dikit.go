package dikit

import (
	"context"
	"net"
	"net/http"

	"log/slog"

	"github.com/ensoria/mb/pkg/mb"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// gRPCサービス登録用のインターフェース
type GRPCServiceRegistrar interface {
	RegisterWithServer(*grpc.Server)
}

var constructors = []any{}

type LC = fx.Lifecycle

func AppendConstructors(adding []any) error {
	constructors = append(constructors, adding...)
	return nil
}

func Constructors() []any {
	return constructors
}

func AsHTTPModule(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"http_modules"`),
	)
}

func AsWSModule(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"ws_modules"`),
	)
}

func AsHTTPPipeline(f any) any {
	return fx.Annotate(f, fx.ParamTags(`group:"http_modules"`))
}

func AsWSRouter(f any) any {
	return fx.Annotate(f, fx.ParamTags(`group:"ws_modules"`))
}

func AsGRPCService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(GRPCServiceRegistrar)),
		fx.ResultTags(`group:"grpc_services"`),
	)
}

func Bind[T any](concrete any) any {
	return fx.Annotate(concrete, fx.As(new(T)))
}

// FIXME: これは、引数が1つの時しか使えない。そのため、基本的にgRPCのときのみなので
// 関数名を別のに変える
// InjectNamedとセットで使う - 結果に名前を付けて提供
func ProvideNamed(constructor any, tag string) any {
	return fx.Annotate(constructor, fx.ResultTags(`name:"`+tag+`"`))
}

// FIXME: これは、引数が1つの時しか使えない。そのため、基本的にgRPCのときのみなので
// 関数名を別のに変える
// ProvideNamedとセットで使う - 名前で依存を注入
func InjectNamed(constructor any, tag string) any {
	return fx.Annotate(constructor, fx.ParamTags(`name:"`+tag+`"`))
}

func ProvideAndRun(constructors []any, invocation any, outputFxLog bool) {
	options := []fx.Option{
		fx.Provide(
			constructors...,
		),
		fx.Invoke(invocation),
	}

	if !outputFxLog {
		options = append(options, fx.NopLogger)
	}

	fx.New(options...).Run()
}

func RegisterGRPCServices() any {
	return fx.Annotate(
		func(
			httpSrv *http.Server,
			grpcSrv *grpc.Server,
			grpcServices []GRPCServiceRegistrar,
		) {
			// gRPCサービスの一括登録
			if grpcSrv != nil {
				for _, service := range grpcServices {
					service.RegisterWithServer(grpcSrv)
				}
				slog.Info("gRPC services registered", "count", len(grpcServices))
			}
		},
		// 第1引数(httpSrv)と第2引数(grpcSrv)にタグは不要なので
		// ``にしてある?
		fx.ParamTags(``, ``, `group:"grpc_services"`),
	)
}

func RegisterHTTPServerLifecycle(lc LC, srv *http.Server) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				slog.Info("HTTP server starting", "addr", srv.Addr)
				if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					slog.Error("HTTP server ListenAndServe error", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Shutting down HTTP server")
			return srv.Shutdown(ctx)
		},
	})
}

func RegisterGRPCServerLifecycle(lc LC, grpcSrv *grpc.Server) {
	if grpcSrv == nil {
		return
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				// TODO: ポートを設定可能にする
				listen, err := net.Listen("tcp", ":50051")
				if err != nil {
					slog.Error("gRPC server failed to listen", "error", err)
					return
				}
				slog.Info("gRPC server starting", "addr", ":50051")
				if err := grpcSrv.Serve(listen); err != nil {
					slog.Error("gRPC server failed to start", "error", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			slog.Info("Shutting down gRPC server")
			grpcSrv.GracefulStop()
			return nil
		},
	})
}

func RegisterMBSubscriberLifecycle(lc LC, subConn mb.Subscriber) {
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			logkit.Info("Shutting down MB subscriber")
			return subConn.Close()
		},
	})
}
