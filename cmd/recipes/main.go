package main

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
	"sync"
	"time"

	"github.com/cdriehuys/recipes"
	"github.com/cdriehuys/recipes/internal/config"
	"github.com/cdriehuys/recipes/migrations"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func serve(ctx context.Context, app application) error {
	srv := &http.Server{
		Addr:     app.config.BindAddr,
		Handler:  app.routes(),
		ErrorLog: slog.NewLogLogger(app.logger.Handler(), slog.LevelError),

		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,

		BaseContext: func(net.Listener) context.Context { return ctx },
	}

	serverErrors := make(chan error)

	var serverWG sync.WaitGroup

	serverWG.Add(1)
	go func() {
		defer serverWG.Done()

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}

		app.logger.Info("Server stopped listening to requests.")
	}()

	go func() {
		<-ctx.Done()

		app.logger.Info("Server shutting down.")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			serverErrors <- err
		}

		// We don't want to close the error channel until the server routine is done using it.
		serverWG.Wait()
		close(serverErrors)

		app.logger.Info("Server shut down.")
	}()

	// Block until the `serverErrors` channel has an error value or closes. If it closes, the
	// received value will be `nil`.
	return <-serverErrors
}

func run(
	ctx context.Context,
	logStream io.Writer,
	args []string,
	migrationFS fs.FS,
	staticFS fs.FS,
	templateFS fs.FS,
) error {
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt)
	defer cancel()

	cmd := &cobra.Command{
		Use:   "recipes",
		Short: "Recipe hosting web application",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			config, err := config.FromEnvironment()
			if err != nil {
				return err
			}

			app, err := newApplication(ctx, logStream, config, migrationFS, staticFS, templateFS)
			if err != nil {
				return err
			}

			return serve(ctx, *app)
		},
	}

	cmd.Flags().String("address", ":8000", "Address to bind the web server to")
	viper.BindPFlag("bind-address", cmd.Flags().Lookup("address"))

	cmd.Flags().Bool("migrate", false, "Run database migrations at startup")
	viper.BindPFlag("run-migrations", cmd.Flags().Lookup("migrate"))

	cmd.SetArgs(args)

	return cmd.ExecuteContext(ctx)
}

func main() {
	err := run(
		context.Background(),
		os.Stderr,
		os.Args[1:],
		migrations.MigrationsFS,
		recipes.StaticFS,
		recipes.TemplateFS,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
	}
}
