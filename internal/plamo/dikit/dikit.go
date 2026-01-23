package dikit

import (
	"context"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

const GroupTagHttpModules = `group:"http_modules"`
const GroupTagWSModules = `group:"ws_modules"`
const GroupTagGRPCServices = `group:"grpc_services"`
const GroupTagWorkerJobs = `group:"worker_jobs"`
const GroupTagScheduledTasks = `group:"scheduled_tasks"`

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
		fx.ResultTags(GroupTagHttpModules),
	)
}

func AsWSModule(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(GroupTagWSModules),
	)
}

func AsGRPCService(f any) any {
	return fx.Annotate(
		f,
		fx.As(new(GRPCServiceRegistrar)),
		fx.ResultTags(GroupTagGRPCServices),
	)
}

func AsWorkerJob(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(GroupTagWorkerJobs),
	)
}

func AsScheduledTask(f any) any {
	return fx.Annotate(
		f,
		fx.ResultTags(GroupTagScheduledTasks),
	)
}

// === Injectors ===

// 汎用版 - 複数の引数位置に対してタグを指定可能
// 例:
// dikit.InjectWithTags(SomeConstructor, `name:"Something"`, `group:"items"`),
func InjectWithTags(constructor any, tags ...string) any {
	return fx.Annotate(constructor, fx.ParamTags(tags...))
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

// TODO: lifecycleもsubscriberとpublisherの登録でしか使われないので、ここでは無くてもいいかも
// 名前も、Subscriberに特化したものだと分かるように変える
func RegisterOnStartLifecycle(lc LC, onStart func(ctx context.Context) error) {
	lc.Append(Hook{
		OnStart: onStart,
	})
}

// こっちも名前を変える。他でも使われてるが、Publisher以外で使われているところは、
// この関数を使わず、そのままlcを使う
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
