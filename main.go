package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

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

func dbTest(logger *slog.Logger, pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var message string
		if err := pool.QueryRow(r.Context(), "SELECT 'Hello, database!'").Scan(&message); err != nil {
			logger.Error("Failed to get database message.", "error", err)
			return
		}

		io.WriteString(w, message)
	}
}

func dbConnectionURL() url.URL {
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	host := os.Getenv("POSTGRES_HOSTNAME")
	db := os.Getenv("POSTGRES_DB")

	return url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(user, password),
		Host:   host,
		Path:   db,
	}
}

func main() {
	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &logOpts))

	connURL := dbConnectionURL()
	dbpool, err := pgxpool.New(context.Background(), connURL.String())
	if err != nil {
		logger.Error("Unable to create database connection pool.", "error", err)
		os.Exit(1)
	}
	defer dbpool.Close()
	logger.Info("Created database connection pool.")

	templateEngine := templates.DiskTemplateEngine{
		IncludePath: "templates/includes",
		LayoutPath:  "templates/layouts",
		Logger:      logger,
	}

	staticServer := staticfiles.StaticFilesFromDisk{BasePath: "static"}

	logger.Info("Creating request handlers.")
	handler := http.NewServeMux()
	handler.HandleFunc("/", index(logger, &templateEngine))
	handler.HandleFunc("/db", dbTest(logger, dbpool))
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
