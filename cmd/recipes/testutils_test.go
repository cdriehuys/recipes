package main

import (
	"bytes"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/cdriehuys/recipes"
	"github.com/cdriehuys/recipes/internal/assert"
	"github.com/cdriehuys/recipes/internal/models/mock"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/templates"
	"github.com/justinas/alice"
	"github.com/neilotoole/slogt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

func (app *application) testAuthenticatePost(w http.ResponseWriter, r *http.Request) {
	userID := r.PostFormValue("userID")
	app.sessionManager.Put(r.Context(), sessionKeyUserID, userID)

	w.WriteHeader(http.StatusNoContent)
}

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
		categoryModel:  &mock.CategoryModel{},
		recipeModel:    &mock.RecipeModel{},
		userModel:      &mock.UserModel{},
		sessionManager: sessionManager,
		staticServer:   &staticServer,
		templates:      &templateWriter,
	}
}

// testServer provides additional utilities on top of the built in test server.
type testServer struct {
	*httptest.Server
}

// newTestServer constructs a test server for the provided application.
func newTestServer(t *testing.T, app *application) *testServer {
	// Add test-only routes
	testMiddleware := alice.New(app.sessionManager.LoadAndSave)

	mux := http.NewServeMux()
	mux.Handle("POST /test/authenticate", testMiddleware.ThenFunc(app.testAuthenticatePost))
	mux.Handle("/", app.routes())

	ts := httptest.NewServer(mux)

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

func (ts *testServer) postForm(t *testing.T, urlPath string, form url.Values) (int, http.Header, string) {
	rs, err := ts.Client().PostForm(ts.URL+urlPath, form)
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

func (ts *testServer) authenticate(t *testing.T, userID string) {
	form := url.Values{}
	form.Add("userID", userID)

	status, _, _ := ts.postForm(t, "/test/authenticate", form)

	assert.Equal(t, http.StatusNoContent, status)
}

func assertRedirects(t *testing.T, headers http.Header, to string) {
	location := headers.Get("Location")
	assert.Equal(t, to, location)
}

func assertLoginRedirect(t *testing.T, headers http.Header, to string) {
	params := url.Values{}
	params.Add("next", to)

	expectedRedirect := "/auth/login?" + params.Encode()

	assertRedirects(t, headers, expectedRedirect)
}

var csrfTokenRX = regexp.MustCompile(`<input type='hidden' name='csrf_token' value='(.+)'>`)

func extractCSRFToken(t *testing.T, body string) string {
	matches := csrfTokenRX.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatal("no csrf token found in body")
	}

	return html.UnescapeString(string(matches[1]))
}
