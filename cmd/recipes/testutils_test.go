package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cdriehuys/recipes"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/templates"
	"github.com/neilotoole/slogt"
)

func newTestApp(t *testing.T) *application {
	logger := slogt.New(t)

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
