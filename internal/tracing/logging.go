package tracing

import (
	"context"
	"log/slog"
)

type Handler struct {
	slog.Handler
}

func (h Handler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(traceCtxKey).(string); ok {
		r.Add("trace_id", traceID)
	}

	return h.Handler.Handle(ctx, r)
}
