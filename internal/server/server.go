package server

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/recipes/internal/config"
	"github.com/cdriehuys/recipes/internal/routes"
	"github.com/gorilla/csrf"
)

type TemplateWriter interface {
	Write(io.Writer, string, map[string]any) error
}

type WebTemplateWriter struct {
	wraps    TemplateWriter
	baseData func(*http.Request) map[string]any
}

func (t WebTemplateWriter) Write(w io.Writer, r *http.Request, name string, data map[string]any) error {
	templateData := t.baseData(r)

	// Make sure caller data can overwrite base template data.
	for k, v := range data {
		templateData[k] = v
	}

	return t.wraps.Write(w, name, templateData)
}

func NewServer(
	logger *slog.Logger,
	config *config.Config,
	templates TemplateWriter,
	oauthConfig routes.OAuthConfig,
	recipeStore routes.RecipeStore,
	userStore routes.UserStore,
	staticServer http.Handler,
) http.Handler {
	webTemplates := WebTemplateWriter{
		wraps: templates,
		baseData: func(r *http.Request) map[string]any {
			return map[string]any{
				csrf.TemplateTag: csrf.TemplateField(r),
			}
		},
	}

	handler := http.NewServeMux()
	routes.AddRoutes(handler, logger, oauthConfig, recipeStore, userStore, webTemplates)

	handler.Handle("/static/", http.StripPrefix("/static/", staticServer))

	return csrf.Protect(config.SecretKey)(handler)
}
