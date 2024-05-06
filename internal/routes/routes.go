package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"

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
	userStore UserStore,
	templates TemplateWriter,
) {
	mux.Handle("GET /{$}", indexHandler(logger, templates))
	mux.Handle("GET /auth/callback", oauthCallbackHandler(logger, oauthConfig, userStore))
	mux.Handle("GET /auth/complete-registration", registerHandler(logger, templates))
	mux.Handle("POST /auth/complete-registration", registerFormHandler(logger, userStore, templates))
	mux.Handle("GET /auth/login", loginHandler(logger, oauthConfig))
	mux.Handle("GET /new-recipe", addRecipeHandler(logger, templates))
	mux.Handle("POST /new-recipe", addRecipeFormHandler(logger, recipeStore, templates))
	mux.Handle("GET /recipes", listRecipeHandler(logger, recipeStore, templates))
	mux.Handle("GET /recipes/{recipeID}", getRecipeHandler(logger, recipeStore, templates))
}

func startRequestLogger(req *http.Request, parent *slog.Logger) *slog.Logger {
	logger := parent.With(slog.Group("request", "method", req.Method, "path", req.URL.Path))
	logger.Info("Handling request.")

	return logger
}
