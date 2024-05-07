package runtime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/cdriehuys/recipes/internal/config"
	"github.com/cdriehuys/recipes/internal/routes"
	"github.com/cdriehuys/recipes/internal/server"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/cdriehuys/recipes/internal/templates"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type staticServer interface {
	http.Handler
	templates.StaticFileFinder
}

func serveHTTP(
	ctx context.Context,
	logger *slog.Logger,
	config *config.Config,
	oauthConfig routes.OAuthConfig,
	recipeStore routes.RecipeStore,
	sessionStore routes.SessionStore,
	userStore routes.UserStore,
	templateEngine server.TemplateWriter,
	staticServer http.Handler,
) error {
	svr := server.NewServer(
		logger,
		config,
		templateEngine,
		oauthConfig,
		recipeStore,
		sessionStore,
		userStore,
		staticServer,
	)
	httpServer := http.Server{
		Addr:              config.BindAddr,
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

func Run(ctx context.Context, logOutput io.Writer, args []string, staticFS, templateFS fs.FS) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cmd := &cobra.Command{
		Use:   "recipes",
		Short: "Recipe hosting web application",
	}

	cmd.Flags().String("address", ":8000", "Address to bind the web server to")
	viper.BindPFlag("bind-address", cmd.Flags().Lookup("address"))

	cmd.Flags().String("secret-key", "", "Secret key")
	viper.BindPFlag("secret-key", cmd.Flags().Lookup("secret-key"))

	cmd.Flags().String("encryption-key", "", "Encryption key")
	viper.BindPFlag("encryption-key", cmd.Flags().Lookup("encryption-key"))

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		config, err := config.FromEnvironment()
		if err != nil {
			return fmt.Errorf("failed to build config: %w", err)
		}

		logOpts := slog.HandlerOptions{
			Level: slog.LevelDebug,
		}
		logger := slog.New(slog.NewTextHandler(logOutput, &logOpts))

		oauthConfig := oauth2.Config{
			ClientID:     config.GoogleClientID,
			ClientSecret: config.GoogleClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  config.OAuthCallbackURL,
			Scopes:       []string{"openid"},
		}

		connURL := config.Database.ConnectionURL()
		dbpool, err := pgxpool.New(ctx, connURL.String())
		if err != nil {
			return fmt.Errorf("unable to create database connection pool: %w", err)
		}
		defer dbpool.Close()
		logger.Info("Created database connection pool.")

		recipeStore := stores.NewRecipeStore(dbpool)
		userStore := stores.NewUserStore(dbpool)

		sessionStore := stores.NewCookieSessionStore(stores.CookieSessionOptions{
			HashKey:       config.SecretKey,
			EncryptionKey: config.EncryptionKey,
			Secure:        false,
			Duration:      7 * 24 * time.Hour,
		})

		var staticServer staticServer
		if config.DevMode {
			logger.Info("Using live static files.")
			staticServer = &staticfiles.StaticFilesFromDisk{BasePath: "static"}
		} else {
			logger.Info("Using precompiled static files.")
			static, err := staticfiles.NewHashedStaticFiles(logger, staticFS, "/static/")
			if err != nil {
				return fmt.Errorf("failed to collect static files: %w", err)
			}

			staticServer = &static
		}

		customFuncs := templates.CustomFunctionMap(staticServer)

		var templateEngine server.TemplateWriter
		if config.DevMode {
			logger.Info("Using live reload template engine.")
			templateEngine = &templates.DiskTemplateEngine{
				IncludePath: "templates/includes",
				LayoutPath:  "templates/layouts",
				FuncMap:     customFuncs,
				Logger:      logger,
			}
		} else {
			logger.Info("Using embedded templates.")
			engine, err := templates.NewFSTemplateEngine(templateFS, customFuncs)
			if err != nil {
				return fmt.Errorf("failed to create template engine: %w", err)
			}

			templateEngine = &engine
		}

		if err := serveHTTP(
			ctx,
			logger,
			&config,
			&oauthConfig,
			recipeStore,
			&sessionStore,
			userStore,
			templateEngine,
			staticServer,
		); err != nil {
			return err
		}

		return nil
	}

	cmd.SetArgs(args)

	return cmd.ExecuteContext(ctx)
}
