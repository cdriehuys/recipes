package tracing

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const traceCtxKey = contextKey("recipes.trace")

// AddTrace adds a unique tracing ID to the request's context before passing it to the next handler.
func AddTrace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := uuid.New().String()
		ctx := context.WithValue(r.Context(), traceCtxKey, traceID)

		tracedRequest := r.WithContext(ctx)

		next.ServeHTTP(w, tracedRequest)
	})
}
