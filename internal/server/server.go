package server

import (
	"log/slog"
	"net/http"

	"github.com/cdriehuys/recipes/internal/routes"
)

func NewServer(
	logger *slog.Logger,
	templates routes.TemplateWriter,
	oauthConfig routes.OAuthConfig,
	recipeStore routes.RecipeStore,
	userStore routes.UserStore,
	staticServer http.Handler,
) http.Handler {
	handler := http.NewServeMux()
	routes.AddRoutes(handler, logger, oauthConfig, recipeStore, userStore, templates)

	handler.Handle("/static/", http.StripPrefix("/static/", staticServer))

	return handler
}
