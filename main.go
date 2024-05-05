package main

import (
	"context"
	"embed"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

	"github.com/cdriehuys/recipes/internal/server"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/templates"
)

//go:embed templates
var templateFS embed.FS

func init() {
	viper.BindEnv("bind-address", "BIND_ADDRESS")
	viper.SetDefault("bind-address", ":8000")

	viper.BindEnv("database.host", "POSTGRES_HOSTNAME")
	viper.BindEnv("database.name", "POSTGRES_DB")
	viper.BindEnv("database.password", "POSTGRES_PASSWORD")
	viper.BindEnv("database.user", "POSTGRES_USER")

	viper.BindEnv("dev-mode", "DEV_MODE")
	viper.SetDefault("dev-mode", false)
}

func dbConnectionURL() url.URL {
	return url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(viper.GetString("database.user"), viper.GetString("database.password")),
		Host:   viper.GetString("database.host"),
		Path:   viper.GetString("database.name"),
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

	var templateEngine server.TemplateEngine
	if viper.GetBool("dev-mode") {
		logger.Info("Using live reload template engine.")
		templateEngine = &templates.DiskTemplateEngine{
			IncludePath: "templates/includes",
			LayoutPath:  "templates/layouts",
			Logger:      logger,
		}
	} else {
		logger.Info("Using embedded templates.")
		engine, err := templates.NewFSTemplateEngine(templateFS)
		if err != nil {
			logger.Error("Failed to create template engine.", "error", err)
			os.Exit(1)
		}

		templateEngine = &engine
	}

	staticServer := staticfiles.StaticFilesFromDisk{BasePath: "static"}

	state := server.State{
		Db:             dbpool,
		Logger:         logger,
		TemplateEngine: templateEngine,
	}

	logger.Info("Creating request handlers.")
	handler := http.NewServeMux()
	handler.Handle("/", server.MakeHandler(state))
	handler.Handle("/static/", http.StripPrefix("/static/", &staticServer))

	server := http.Server{
		Addr:              viper.GetString("bind-address"),
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
