package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/cdriehuys/recipes/internal/domain"
	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type OAuthConfig interface {
	AuthCodeURL(string, ...oauth2.AuthCodeOption) string
	Client(context.Context, *oauth2.Token) *http.Client
	Exchange(context.Context, string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type SessionStore interface {
	Create(context.Context, http.ResponseWriter, string) error
	IsAuthenticated(*http.Request) bool
	UserID(*http.Request) (string, error)
}

type RecipeStore interface {
	Add(context.Context, *slog.Logger, domain.NewRecipe) error
	GetByID(context.Context, *slog.Logger, uuid.UUID) (stores.Recipe, error)
	List(context.Context, *slog.Logger) ([]stores.RecipeListItem, error)
}

type UserStore interface {
	RecordLogIn(context.Context, *slog.Logger, string) (bool, error)
	UpdateDetails(context.Context, *slog.Logger, string, domain.UserDetails) error
}

type TemplateWriter interface {
	Write(io.Writer, *http.Request, string, map[string]any) error
}

func AddRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	oauthConfig OAuthConfig,
	recipeStore RecipeStore,
	sessionStore SessionStore,
	userStore UserStore,
	templates TemplateWriter,
) {
	authMiddleware := requireAuth(sessionStore)

	mux.Handle("GET /{$}", indexHandler(logger, templates))
	mux.Handle("GET /auth/callback", oauthCallbackHandler(logger, oauthConfig, sessionStore, userStore))
	mux.Handle("GET /auth/complete-registration", registerHandler(logger, templates))
	mux.Handle("POST /auth/complete-registration", registerFormHandler(logger, userStore, templates))
	mux.Handle("GET /auth/login", loginHandler(oauthConfig))
	mux.Handle("GET /new-recipe", authMiddleware(addRecipeHandler(logger, templates)))
	mux.Handle("POST /new-recipe", authMiddleware(addRecipeFormHandler(logger, recipeStore, templates)))
	mux.Handle("GET /recipes", authMiddleware(listRecipeHandler(logger, recipeStore, templates)))
	mux.Handle("GET /recipes/{recipeID}", authMiddleware(getRecipeHandler(logger, recipeStore, templates)))
}

func startRequestLogger(req *http.Request, parent *slog.Logger) *slog.Logger {
	logger := parent.With(slog.Group("request", "method", req.Method, "path", req.URL.Path))
	logger.Info("Handling request.")

	return logger
}

func requireAuth(session SessionStore) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !session.IsAuthenticated(r) {
				loginParams := url.Values{}
				loginParams.Set("next", r.URL.Path)

				login := url.URL{
					Path:     "/auth/login",
					RawQuery: loginParams.Encode(),
				}

				http.Redirect(w, r, login.String(), http.StatusSeeOther)
			} else {
				h.ServeHTTP(w, r)
			}
		})
	}
}
