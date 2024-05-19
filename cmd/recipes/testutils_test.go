package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cdriehuys/recipes"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/templates"
	"github.com/neilotoole/slogt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func newTestApp(t *testing.T) *application {
	logger := slogt.New(t)

	oauthConfig := oauth2.Config{
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		Endpoint:     google.Endpoint,
		RedirectURL:  "https://example.com",
		Scopes:       []string{"openid"},
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour

	staticServer, err := staticfiles.NewHashedStaticFiles(logger, recipes.StaticFS, "/static/")
	if err != nil {
		t.Fatal(err)
	}

	templateWriter, err := templates.NewFSTemplateEngine(recipes.TemplateFS, templates.CustomFunctionMap(&staticServer))
	if err != nil {
		t.Fatal(err)
	}

	return &application{
		logger:         logger,
		oauthConfig:    &oauthConfig,
		sessionManager: sessionManager,
		staticServer:   &staticServer,
		templates:      &templateWriter,
	}
}

// testServer provides additional utilities on top of the built in test server.
type testServer struct {
	*httptest.Server
}

// newTestServer constructs a test server that utilizes the provided handler.
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}

	ts.Client().Jar = jar

	// Don't follow redirects so we can assert on expected redirects.
	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) do(t *testing.T, req *http.Request) (int, *http.Response) {
	baseURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}

	// Need to clear the request URI or else the request fails.
	req.RequestURI = ""
	req.URL.Scheme = baseURL.Scheme
	req.URL.Host = baseURL.Host

	rs, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	return rs.StatusCode, rs
}

// Implement a get() method on our custom testServer type. This makes a GET
// request to a given url path using the test server client, and returns the
// response status code, headers and body.
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}

	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	body = bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}
