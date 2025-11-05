package bootstrap

import (
	"github.com/ensoria/config/pkg/env"
	"github.com/ensoria/projecttemplate/internal/plamo/dikit"
	"github.com/ensoria/projecttemplate/internal/plamo/grpckit"
	"github.com/ensoria/projecttemplate/internal/plamo/logkit"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

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
