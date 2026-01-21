package server

import (
	"context"
	"log/slog"
	"net"

	"github.com/ensoria/config/pkg/env"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/grpckit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// gRPCサーバーの初期化
func NewGRPCApp(envVal *string) func(lc dikit.LC, grpcServices []dikit.GRPCServiceRegistrar) *grpc.Server {
	return func(lc dikit.LC, grpcServices []dikit.GRPCServiceRegistrar) *grpc.Server {
		// ログとpanicリカバリinterceptor付きのgRPCサーバーを作成
		grpcSrv := grpckit.NewGRPCServer(logkit.Logger())

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

// REFACTOR: serverに移すか?
// gRPC server lifecycle registration
func RegisterGRPCServerLifecycle(lc dikit.LC, grpcSrv *grpc.Server) {
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
