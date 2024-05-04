package main

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/templates"
)

type TemplateEngine interface {
	Write(w io.Writer, name string, data any) error
}

func index(logger *slog.Logger, templates TemplateEngine) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger.Info("Handling index request.")

		if err := templates.Write(w, "index", nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	}
}

func main() {
	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &logOpts))
	logger.Info("Creating request handlers.")

	templateEngine := templates.DiskTemplateEngine{
		IncludePath: "templates/includes",
		LayoutPath:  "templates/layouts",
		Logger:      logger,
	}

	staticServer := staticfiles.StaticFilesFromDisk{BasePath: "static"}

	handler := http.NewServeMux()
	handler.HandleFunc("/", index(logger, &templateEngine))
	handler.Handle("/static/", http.StripPrefix("/static/", &staticServer))

	server := http.Server{
		Addr:              "0.0.0.0:8000",
		Handler:           handler,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
	}

	logger.Info("Starting server.", "address", server.Addr)
	if err := server.ListenAndServe(); err != nil {
		logger.Error("Server failed.", "error", err)
	}
}
