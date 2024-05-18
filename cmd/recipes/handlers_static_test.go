package main

import (
	"net/http"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

func TestStaticRoutes(t *testing.T) {
	app := newTestApp(t)
	server := newTestServer(t, app.routes())

	staticPages := []string{"/", "/privacy-policy"}
	for _, page := range staticPages {
		t.Run("GET "+page, func(t *testing.T) {
			status, _, _ := server.get(t, page)
			assert.Equal(t, http.StatusOK, status)
		})
	}
}
