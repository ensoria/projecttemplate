package grpckit

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ensoria/grpcgear/pkg/interceptor/logging"
	"github.com/ensoria/grpcgear/pkg/interceptor/recovery/recoverysrv"
)

func CreateBasicPanicLogFuncs(logger logging.Logger) (recoverysrv.LogUnaryPanicFunc, recoverysrv.LogStreamPanicFunc) {

	logUnary := func(info *recoverysrv.UnaryPanicInfo) {
		attrs := buildPanicLog(
			info.Server.FullMethod,
			info.Client,
			fmt.Sprint(info.Value),
			info.ValueTrunc,
			info.Stack,
			info.StackTrunc,
			info.TraceID,
			info.SpanID,
			info.TraceSampled,
		)
		logger.Error("gRPC unary handler panic recovered", attrs...)
	}

	logStream := func(info *recoverysrv.StreamPanicInfo) {
		extra := []any{"client_stream", info.Server.IsClientStream, "server_stream", info.Server.IsServerStream}
		attrs := buildPanicLog(
			info.Server.FullMethod,
			info.Client,
			fmt.Sprint(info.Value),
			info.ValueTrunc,
			info.Stack,
			info.StackTrunc,
			info.TraceID,
			info.SpanID,
			info.TraceSampled,
			extra...,
		)
		logger.Error("gRPC stream handler panic recovered", attrs...)
	}

	return logUnary, logStream
}

// panic 時に一般的に欲しい情報を整理し、順序を安定させる
// 1. 固定基本項目 (method, panic_value, flags, trace/span/request など)
// 2. client メタデータの昇格キー (request_id, trace_id など) は重複排除
// 3. 残りの client メタデータはキーを正規化 ( - -> _ ) してソート
func buildPanicLog(method string, client map[string]any, panicValue string, valueTrunc bool, stack string, stackTrunc bool, traceID, spanID string, sampled bool, extra ...any) []any {
	attrs := []any{
		"method", method,
		"panic_value", panicValue,
	}
	if valueTrunc {
		attrs = append(attrs, "panic_value_truncated", true)
	}
	if stack != "" { // stack は設定で抑制される可能性がある
		attrs = append(attrs, "stack_trace", stack)
		if stackTrunc {
			attrs = append(attrs, "stack_truncated", true)
		}
	}
	// Trace 情報 (存在すれば)
	if traceID != "" {
		attrs = append(attrs, "trace_id", traceID)
	}
	if spanID != "" {
		attrs = append(attrs, "span_id", spanID)
	}
	if sampled {
		attrs = append(attrs, "trace_sampled", true)
	}

	// 昇格キー
	promote := []string{
		"request_id", "retry_attempt", "client_ip", "client_port", "network", "auth_type",
		"deadline_remaining_ms", "deadline_unix",
	}
	seen := map[string]struct{}{}
	// trace_id / span_id / trace_sampled はすでに追加済みのため除外
	for _, k := range promote {
		if v, ok := client[k]; ok {
			attrs = append(attrs, k, v)
			seen[k] = struct{}{}
		}
	}
	// user-agent -> user_agent
	if ua, ok := client["user-agent"]; ok {
		attrs = append(attrs, "user_agent", ua)
		seen["user-agent"] = struct{}{}
	}

	// 残りをソートして付与 (重複と基本キー除外)
	remaining := make([]string, 0, len(client))
	for k := range client {
		if _, skip := seen[k]; skip {
			continue
		}
		remaining = append(remaining, k)
	}
	// method / panic_value / stack_trace などは除外対象
	exclude := map[string]struct{}{"method": {}, "panic_value": {}, "stack_trace": {}, "trace_id": {}, "span_id": {}, "trace_sampled": {}}
	sort.Strings(remaining)
	for _, k := range remaining {
		if _, skip := exclude[k]; skip {
			continue
		}
		v := client[k]
		ck := strings.ReplaceAll(k, "-", "_")
		if ck == "user_agent" { // 既に追加済み
			continue
		}
		attrs = append(attrs, ck, v)
	}

	if len(extra) > 0 {
		attrs = append(attrs, extra...)
	}
	return attrs
}
