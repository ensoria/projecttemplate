package dikit

import (
	"context"
	"net"
	"net/http"

	"log/slog"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// gRPCサービス登録用のインターフェース
type GRPCServiceRegistrar interface {
	RegisterWithServer(*grpc.Server)
}

type LC = fx.Lifecycle

var constructors = []any{}

func AppendConstructors(adding []any) {
	constructors = append(constructors, adding...)
}

func Constructors() []any {
	return constructors
}

var invocations = []any{}

func AppendInvocations(adding []any) {
	invocations = append(invocations, adding...)
}

func Invocations() []any {
	return invocations
}

// === Providers ===

// Tのインターフェースに対して、該当するconcreteが1つだけの場合に使う
func ProvideAs[T any](concrete any) any {
	return fx.Annotate(concrete, fx.As(new(T)))
}

// Tのインターフェースに対して、該当するconcreteが複数ある場合に、
// タグ付きで登録することで注入する際にどのconcreteかを指定できる
func ProvideAsNamed[T any](concrete any, tag string) any {
	return fx.Annotate(concrete, fx.As(new(T)), fx.ResultTags(`name:"`+tag+`"`))
}

// TODO: ProvideAsNamedとの違いは?
func ProvideNamed(constructor any, tag string) any {
	return fx.Annotate(constructor, fx.ResultTags(`name:"`+tag+`"`))
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

func AsGRPCService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(GRPCServiceRegistrar)),
		fx.ResultTags(`group:"grpc_services"`),
	)
}

// === Injectors ===

// 汎用版 - 複数の引数位置に対してタグを指定可能
// 例:
// dikit.InjectWithTags(SomeConstructor, “, `name:"Something"`, `group:"items"`),
func InjectWithTags(constructor any, tags ...string) any {
	return fx.Annotate(constructor, fx.ParamTags(tags...))
}

func InjectHTTPModules(f any) any {
	return fx.Annotate(f, fx.ParamTags(`group:"http_modules"`))
}

func InjectWSModules(f any) any {
	return fx.Annotate(f, fx.ParamTags(`group:"ws_modules"`))
}

// Subscriber注入用
func InjectSubscriber(constructor any, tag string) any {
	return fx.Annotate(
		constructor,
		fx.ParamTags(``, ``, `name:"`+tag+`"`),
	)
}

// gRPCクライアントの注入用
// 実際には引数が1つだけの場合は汎用的に使えますが、
// 汎用的に使いたい場合は、別の関数を用意するか、IbjectWithTagsを使ってください。
func InjectGRPCClient(constructor any, tag string) any {
	return fx.Annotate(constructor, fx.ParamTags(`name:"`+tag+`"`))
}

// === invocations ===

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

// HTTP/WebSocket controllers lifecycle registration
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

// gRPC server lifecycle registration
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

func RegisterOnStartLifecycle(lc LC, onStart func(ctx context.Context) error) {
	lc.Append(fx.Hook{
		OnStart: onStart,
	})
}

func RegisterOnStopLifecycle(lc LC, onStop func(ctx context.Context) error) {
	lc.Append(fx.Hook{
		OnStop: onStop,
	})
}

// 汎用的なfxアプリケーションの提供と実行

func ProvideAndRun(constructors []any, invocations []any, outputFxLog bool) {
	options := []fx.Option{
		fx.Provide(
			constructors...,
		),
		fx.Invoke(invocations...),
	}

	if !outputFxLog {
		options = append(options, fx.NopLogger)
	}

	fx.New(options...).Run()
}
