package routes

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/recipes/internal/domain"
	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/google/uuid"
)

type RecipeStore interface {
	Add(context.Context, *slog.Logger, domain.NewRecipe) error
	GetByID(context.Context, *slog.Logger, uuid.UUID) (stores.Recipe, error)
	List(context.Context, *slog.Logger) ([]stores.RecipeListItem, error)
}

type TemplateWriter interface {
	Write(io.Writer, string, any) error
}

func AddRoutes(
	mux *http.ServeMux,
	logger *slog.Logger,
	recipeStore RecipeStore,
	templates TemplateWriter,
) {
	mux.Handle("GET /{$}", indexHandler(logger, templates))
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
