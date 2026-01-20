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
type Hook = fx.Hook

var constructors = []any{}

// Constructorとして登録した関数は、参照されて初めて実行されます。
// 参照されていなくても、必ず実行してほしい関数は、AppendInvocationsを使って
// 登録してください。
// 登録するconstructor関数は、戻り値が必須です
func AppendConstructors(adding []any) {
	constructors = append(constructors, adding...)
}

func Constructors() []any {
	return constructors
}

var invocations = []any{}

// Invocationは、アプリ起動時に必ず実行されるものです。
// Constructorとは違い、参照されていなくても実行されます。
// 参照されていなくても必ず実行してほしい関数は、ここに登録してください。
// 登録するinvocation関数は戻り値は必須ではありません。
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

// TODO:
func AsScheduledTask(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(`group:"scheduled_tasks"`),
	)
}

// === Injectors ===

// 汎用版 - 複数の引数位置に対してタグを指定可能
// 例:
// dikit.InjectWithTags(SomeConstructor, `name:"Something"`, `group:"items"`),
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

// DELETE: この関数は不要になる?必要だとしても、引数を2番目に設定する必要があるか?
func InjectGRPCServices(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(`group:"grpc_services"`),
	)
}

// gRPCクライアントの注入用
// 実際には引数が1つだけの場合は汎用的に使えますが、
// 汎用的に使いたい場合は、別の関数を用意するか、IbjectWithTagsを使ってください。
func InjectGRPCClient(constructor any, tag string) any {
	return fx.Annotate(constructor, fx.ParamTags(`name:"`+tag+`"`))
}

func InjectScheduledTasks(f any) any {
	return fx.Annotate(
		f,
		fx.ParamTags(``, ``, `group:"scheduled_tasks"`),
	)
}

// REFACTOR: serverに移すか?
// HTTP/WebSocket controllers lifecycle registration
func RegisterHTTPServerLifecycle(lc LC, srv *http.Server) {
	lc.Append(Hook{
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

// REFACTOR: serverに移すか?
// gRPC server lifecycle registration
func RegisterGRPCServerLifecycle(lc LC, grpcSrv *grpc.Server) {
	if grpcSrv == nil {
		return
	}

	lc.Append(Hook{
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
	lc.Append(Hook{
		OnStart: onStart,
	})
}

func RegisterOnStopLifecycle(lc LC, onStop func(ctx context.Context) error) {
	lc.Append(Hook{
		OnStop: onStop,
	})
}

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
