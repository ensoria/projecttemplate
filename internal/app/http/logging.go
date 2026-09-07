// This file holds the records the global middlewares write. They live beside
// the chain that installs them rather than inside its declaration, so that the
// list a project edits stays a list.
//
// Both records carry a stable "type", which is what an alert condition matches
// on; the values are listed in the README's alerting section.

package http

import (
	"fmt"
	"log/slog"

	"github.com/ensoria/ensoria-template/internal/plamo/restkit"
	"github.com/ensoria/loggear/pkg/loggear"
	"github.com/ensoria/rest/pkg/rest"
)

func logIncomingRequest(req *rest.Request, res *rest.Response) {
	loggear.Info("HTTP Request",
		slog.String("method", req.Method()),
		slog.String("path", req.Path()),
		slog.Int("status_code", res.Code),
		slog.String("remote_addr", req.RemoteAddr()),
		slog.String("user_agent", req.UserAgent()),
		slog.String("type", "access_log"),
	)
}

// panicViolationGroup is the key a contract violation's fields are nested under
// in a panic record.
//
// They are grouped rather than written flat beside the record's own fields
// because a violation names the endpoint as well: expanded flat, "method" would
// appear in one JSON object twice and neither occurrence could be relied on.
// Grouping also leaves "type" saying panic_log and nothing else, which is what a
// panic alert matches on.
const panicViolationGroup = "contract_violation"

// logPanicDetails records a panic the recovery middleware caught. It is the
// only account of a request that failed for a reason nobody anticipated, so it
// carries what identifies the defect — the value, its type, the stack — and the
// least that identifies where it came from: the endpoint and the caller's
// address, which is what makes it possible to see what else that caller sent.
//
// It deliberately leaves out the client's user agent. A panic is a defect in
// this server, and which browser asked does not help say which defect it is;
// the access log written for the same request carries it for the cases where
// the client population is the question.
func logPanicDetails(r *rest.Request, panicValue interface{}, stackTrace []byte) {
	args := []any{
		slog.String("method", r.Method()),
		slog.String("url", r.URLStr()),
		slog.String("remote_addr", r.RemoteAddr()),
		slog.Any("panic_value", panicValue),
		slog.String("panic_type", fmt.Sprintf("%T", panicValue)),
		slog.String("stack_trace", string(stackTrace)),
		slog.String("type", "panic_log"),
	}
	// A panic carrying a contract violation can say which promise the
	// implementation broke, which a stack trace of generic frames cannot. The
	// assertion is on the interface rather than on any one kind of violation, so
	// a kind added later is expanded here without this file being touched.
	if violation, ok := panicValue.(restkit.ContractViolation); ok {
		args = append(args, slog.Group(panicViolationGroup, restkit.LogArgs(violation)...))
	}
	loggear.Error("Panic Recovered", args...)
}
