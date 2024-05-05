package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/viper"

	"github.com/cdriehuys/recipes/internal/routes"
	"github.com/cdriehuys/recipes/internal/server"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/stores"
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

func serveHTTP(
	ctx context.Context,
	logger *slog.Logger,
	recipeStore routes.RecipeStore,
	templateEngine routes.TemplateWriter,
	staticServer http.Handler,
) error {
	svr := server.NewServer(logger, templateEngine, recipeStore, staticServer)
	httpServer := http.Server{
		Addr:              viper.GetString("bind-address"),
		Handler:           svr,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       5 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	serverErrors := make(chan error)

	// Start the server in a separate go-routine so we can listen for cancellation signals and shut
	// down gracefully.
	go func() {
		logger.Info("Starting server.", "address", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	waitChan := make(chan struct{})

	go func() {
		defer close(waitChan)
		<-ctx.Done()

		logger.Info("Gracefully shutting down server.")

		// Create a new context to limit the shutdown time for the server. It should already be
		// closing due to the cancellation of the server's original context.
		shutdownCtx := context.Background()
		shutdownCtx, cancel := context.WithTimeout(shutdownCtx, 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			serverErrors <- err
		}
	}()

	select {
	case err := <-serverErrors:
		return err
	case <-waitChan:
		logger.Info("Server shut down normally.")
		return nil
	}
}

func run(ctx context.Context, logOutput io.Writer) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	logOpts := slog.HandlerOptions{
		Level: slog.LevelDebug,
	}
	logger := slog.New(slog.NewTextHandler(logOutput, &logOpts))

	connURL := dbConnectionURL()
	dbpool, err := pgxpool.New(ctx, connURL.String())
	if err != nil {
		return fmt.Errorf("unable to create database connection pool: %w", err)
	}
	defer dbpool.Close()
	logger.Info("Created database connection pool.")

	recipeStore := stores.NewRecipeStore(dbpool)

	var templateEngine routes.TemplateWriter
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
			return fmt.Errorf("failed to create template engine: %w", err)
		}

		templateEngine = &engine
	}

	staticServer := staticfiles.StaticFilesFromDisk{BasePath: "static"}

	if err := serveHTTP(ctx, logger, recipeStore, templateEngine, &staticServer); err != nil {
		return err
	}

	return nil
}

func main() {
	// Delegate to `run` which can return errors instead of having to worry about exit codes. This
	// also provides concrete injections of global state like output streams to make testing easier.
	ctx := context.Background()
	if err := run(ctx, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
