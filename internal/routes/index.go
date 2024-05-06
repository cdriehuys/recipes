package routes

import (
	"log/slog"
	"net/http"
)

func indexHandler(logger *slog.Logger, templates TemplateWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		if err := templates.Write(w, r, "index", nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	})
}
