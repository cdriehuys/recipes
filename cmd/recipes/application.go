package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/pgxstore"
	"github.com/alexedwards/scs/v2"
	"github.com/cdriehuys/recipes/internal/config"
	"github.com/cdriehuys/recipes/internal/domain"
	"github.com/cdriehuys/recipes/internal/staticfiles"
	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/cdriehuys/recipes/internal/templates"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

type oauthConfig interface {
	AuthCodeURL(string, ...oauth2.AuthCodeOption) string
	Client(context.Context, *oauth2.Token) *http.Client
	Exchange(context.Context, string, ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

type recipeStore interface {
	Add(context.Context, *slog.Logger, stores.Recipe) error
	GetByID(context.Context, *slog.Logger, string, uuid.UUID) (stores.Recipe, error)
	List(context.Context, *slog.Logger, string) ([]stores.Recipe, error)
	Update(context.Context, stores.Recipe) error
}

type userStore interface {
	Exists(context.Context, string) (bool, error)
	RecordLogIn(context.Context, *slog.Logger, string) (bool, error)
	UpdateDetails(context.Context, *slog.Logger, string, domain.UserDetails) error
}

type templateWriter interface {
	Write(io.Writer, *http.Request, string, any) error
}

type staticServer interface {
	http.Handler
	templates.StaticFileFinder
}

type application struct {
	logger         *slog.Logger
	config         config.Config
	oauthConfig    oauthConfig
	recipeStore    recipeStore
	userStore      userStore
	templates      templateWriter
	sessionManager *scs.SessionManager
	staticServer   staticServer
}

func newApplication(
	ctx context.Context,
	logStream io.Writer,
	config config.Config,
	migrationFS fs.FS,
	staticFS fs.FS,
	templateFS fs.FS,
) (*application, error) {
	logger := slog.New(slog.NewTextHandler(logStream, nil))

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
		return nil, fmt.Errorf("unable to create database connection pool: %w", err)
	}
	logger.Info("Created database connection pool.")

	if config.RunMigrations {
		logger.Info("Running database migrations.")
		runner := func(conn *pgxpool.Conn) error {
			return runMigrations(ctx, conn.Conn(), migrationFS)
		}

		if err := dbpool.AcquireFunc(ctx, runner); err != nil {
			return nil, fmt.Errorf("failed to run migrations: %w", err)
		}

		logger.Info("Database migration ran successfully.")
	}

	sessionManager := scs.New()
	sessionManager.Store = pgxstore.New(dbpool)

	recipeStore := stores.NewRecipeStore(dbpool)
	userStore := stores.NewUserStore(dbpool)

	var staticServer staticServer
	if config.DevMode {
		logger.Info("Using live static files.")
		staticServer = &staticfiles.StaticFilesFromDisk{BasePath: "static"}
	} else {
		logger.Info("Using precompiled static files.")
		static, err := staticfiles.NewHashedStaticFiles(logger, staticFS, "/static/")
		if err != nil {
			return nil, fmt.Errorf("failed to collect static files: %w", err)
		}

		staticServer = &static
	}

	customFuncs := templates.CustomFunctionMap(staticServer)

	var templateEngine templateWriter
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
			return nil, fmt.Errorf("failed to create template engine: %w", err)
		}

		templateEngine = &engine
	}

	app := &application{
		logger:         logger,
		config:         config,
		oauthConfig:    &oauthConfig,
		recipeStore:    recipeStore,
		userStore:      userStore,
		templates:      templateEngine,
		sessionManager: sessionManager,
		staticServer:   staticServer,
	}

	return app, nil
}

func runMigrations(ctx context.Context, conn *pgx.Conn, migrations fs.FS) error {
	migrator, err := migrate.NewMigrator(ctx, conn, "public.schema_version")
	if err != nil {
		return err
	}

	if err := migrator.LoadMigrations(migrations); err != nil {
		return err
	}

	return migrator.Migrate(ctx)
}
