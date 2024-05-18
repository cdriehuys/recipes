package main

import (
	"net/http"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

func TestIndex(t *testing.T) {
	app := newTestApp(t)
	server := newTestServer(t, app.routes())

	status, _, _ := server.get(t, "/")

	assert.Equal(t, http.StatusOK, status)
}
