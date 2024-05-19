package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

func getNonceCookie(t *testing.T, cookies []*http.Cookie) *http.Cookie {
	for _, cookie := range cookies {
		if cookie.Name == "recipes.state" {
			return cookie
		}
	}

	t.Fatalf("No cookie named 'recipes.state': %v", cookies)

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
	app := newTestApp(t)
	server := newTestServer(t, app.routes())

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
	stateCookie := getNonceCookie(t, resp.Cookies())

	assert.Equal(t, nonce, stateCookie.Value)

	// Without the state, the URL should match our expectation
	expectedState := url.Values{}
	expectedState.Add("next", "")
	expectedRedirect := app.oauthConfig.AuthCodeURL(expectedState.Encode())

	receivedRedirect := removeNonce(t, resp.Header.Get("Location"))

	assert.Equal(t, expectedRedirect, receivedRedirect)
}
