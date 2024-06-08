package tracing

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Test_AddTrace(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		traceID, ok := r.Context().Value(traceCtxKey).(string)
		if !ok {
			t.Error("Expected request to have tracing ID")
		}

		if traceID == "" {
			t.Error("Expected trace ID to be non-empty")
		}
	}

	decorated := AddTrace(http.HandlerFunc(handler))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)

	decorated.ServeHTTP(w, r)
}
