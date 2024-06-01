package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

type mockSessionManager struct {
	puts map[string]any
}

func (s *mockSessionManager) GetString(context.Context, string) string {
	return ""
}

func (s *mockSessionManager) LoadAndSave(next http.Handler) http.Handler {
	return next
}

func (s *mockSessionManager) PopString(context.Context, string) string {
	return ""
}

func (s *mockSessionManager) Put(_ context.Context, key string, value any) {
	if s.puts == nil {
		s.puts = make(map[string]any)
	}

	s.puts[key] = value
}

func (s *mockSessionManager) Remove(context.Context, string) {}

func (s *mockSessionManager) RenewToken(context.Context) error {
	return nil
}

func removeNonce(t *testing.T, redirectURL string) string {
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		t.Fatal(err)
	}

	query := parsed.Query()
	state, err := url.ParseQuery(query.Get("state"))
	if err != nil {
		t.Fatal(err)
	}

	state.Del("nonce")
	query.Set("state", state.Encode())
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func TestAuthLogin(t *testing.T) {
	sessionManager := mockSessionManager{}

	app := newTestApp(t)
	app.sessionManager = &sessionManager
	server := newTestServer(t, app)

	req := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
	status, resp := server.do(t, req)
	assert.Equal(t, http.StatusTemporaryRedirect, status)

	redirect, err := url.Parse(resp.Header.Get("Location"))
	if err != nil {
		t.Fatalf("Failed to parse 'Location' header as URL: %v", err)
	}

	redirectValues, err := url.ParseQuery(redirect.RawQuery)
	if err != nil {
		t.Fatal(err)
	}

	state, err := url.ParseQuery(redirectValues.Get("state"))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "", state.Get("next"))

	nonce := state.Get("nonce")
	if err != nil {
		t.Fatal(err)
	}
	sessionNonce := sessionManager.puts["oauth-nonce"].(string)

	assert.Equal(t, nonce, sessionNonce)

	// Without the state, the URL should match our expectation
	expectedState := url.Values{}
	expectedState.Add("next", "")
	expectedRedirect := app.oauthConfig.AuthCodeURL(expectedState.Encode())

	receivedRedirect := removeNonce(t, resp.Header.Get("Location"))

	assert.Equal(t, expectedRedirect, receivedRedirect)
}
