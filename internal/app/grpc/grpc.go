package grpc

import (
	"context"
	"log/slog"
	"net"

	"github.com/ensoria/config/pkg/env"
	"github.com/ensoria/grpcgear/pkg/interceptor/logging"
	"github.com/ensoria/grpcgear/pkg/interceptor/logging/logsrv"
	"github.com/ensoria/grpcgear/pkg/interceptor/recovery/recoverysrv"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	ggrpc "google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// gRPCサーバーの初期化
func NewGRPCApp(envVal *string) func(lc dikit.LC, grpcServices []dikit.GRPCServiceRegistrar) *ggrpc.Server {
	return func(lc dikit.LC, grpcServices []dikit.GRPCServiceRegistrar) *ggrpc.Server {
		// ログとpanicリカバリinterceptor付きのgRPCサーバーを作成
		grpcSrv := NewGRPCServer(logkit.Logger())

		// reflectionは開発環境でのみ有効にする
		// TODO: config/env にIsLocal()を作って、それを使う
		if envVal != nil && (*envVal == string(env.Local) || *envVal == string(env.Develop)) {
			reflection.Register(grpcSrv)
			logkit.Info("gRPC reflection enabled for development environment", "env", *envVal)
		}

		RegisterGRPCServerLifecycle(lc, grpcSrv)

		for _, svc := range grpcServices {
			svc.RegisterWithServer(grpcSrv)
		}
		logkit.Info("gRPC services registered", "count", len(grpcServices))

		return grpcSrv
	}
}

// DELETE: この関数は不要?
func CreateGRPCServices(modules []dikit.GRPCServiceRegistrar) []dikit.GRPCServiceRegistrar {
	return modules
}

func NewGRPCServer(logger logging.Logger) *ggrpc.Server {
	logCfg := LogConfig()
	recCfg := recoverysrv.DefaultRecoveryConfig()
	logUnarySuccess, logUnaryError := CreateBasicUnaryLogFuncs(logger)
	logStreamSuccess, logStreamError := CreateBasicStreamLogFuncs(logger)
	logUnaryPanic, logStreamPanic := CreateBasicPanicLogFuncs(logger)

	// チェーン化された複数のinterceptorを作成
	// 注意: 実行される順番は引数で渡す順番です。
	// そのため、確実にpanicを拾う場合はrecoveryを最初に配置すべきです
	opts := []ggrpc.ServerOption{
		ggrpc.ChainUnaryInterceptor(
			recoverysrv.RecoveryUnaryInterceptor(logUnaryPanic, logCfg, recCfg), // 最外側: panic を最初にキャッチ
			logsrv.LoggingUnaryInterceptor(logUnarySuccess, logUnaryError, logCfg),
		),
		ggrpc.ChainStreamInterceptor(
			recoverysrv.RecoveryStreamInterceptor(logStreamPanic, logCfg, recCfg), // 最外側: panic を最初にキャッチ
			logsrv.LoggingStreamInterceptor(logStreamSuccess, logStreamError, logCfg),
		),
	}

	return ggrpc.NewServer(opts...)
}

// gRPC server lifecycle registration
func RegisterGRPCServerLifecycle(lc dikit.LC, grpcSrv *ggrpc.Server) {
	if grpcSrv == nil {
		return
	}

	lc.Append(dikit.Hook{
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
