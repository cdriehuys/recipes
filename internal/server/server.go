package server

import (
	"log/slog"
	"net/http"

	"github.com/cdriehuys/recipes/internal/routes"
)

func NewServer(
	logger *slog.Logger,
	templates routes.TemplateWriter,
	recipeStore routes.RecipeStore,
	staticServer http.Handler,
) http.Handler {
	handler := http.NewServeMux()
	routes.AddRoutes(handler, logger, recipeStore, templates)

	handler.Handle("/static/", http.StripPrefix("/static/", staticServer))

	return handler
}
