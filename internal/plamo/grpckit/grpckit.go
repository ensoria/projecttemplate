package grpckit

import (
	"github.com/ensoria/grpcgear/pkg/interceptor/logging"
	"github.com/ensoria/grpcgear/pkg/interceptor/logging/logsrv"
	"github.com/ensoria/grpcgear/pkg/interceptor/recovery/recoverysrv"
	"google.golang.org/grpc"
)

// FIXME: grpcサーバーに関する設定がここでいいのか、考え直す

func NewGRPCServer(logger logging.Logger) *grpc.Server {
	logCfg := LogConfig()
	recCfg := recoverysrv.DefaultRecoveryConfig()
	logUnarySuccess, logUnaryError := CreateBasicUnaryLogFuncs(logger)
	logStreamSuccess, logStreamError := CreateBasicStreamLogFuncs(logger)
	logUnaryPanic, logStreamPanic := CreateBasicPanicLogFuncs(logger)

	// チェーン化された複数のinterceptorを作成
	// 注意: 実行される順番は引数で渡す順番です。
	// そのため、確実にpanicを拾う場合はrecoveryを最初に配置すべきです
	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(
			recoverysrv.RecoveryUnaryInterceptor(logUnaryPanic, logCfg, recCfg), // 最外側: panic を最初にキャッチ
			logsrv.LoggingUnaryInterceptor(logUnarySuccess, logUnaryError, logCfg),
		),
		grpc.ChainStreamInterceptor(
			recoverysrv.RecoveryStreamInterceptor(logStreamPanic, logCfg, recCfg), // 最外側: panic を最初にキャッチ
			logsrv.LoggingStreamInterceptor(logStreamSuccess, logStreamError, logCfg),
		),
	}

	return grpc.NewServer(opts...)
}
