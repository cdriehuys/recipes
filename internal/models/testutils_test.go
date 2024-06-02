package models_test

import (
	"context"
	"io/fs"
	"os"
	"testing"

	"github.com/cdriehuys/recipes/internal/config"
	"github.com/cdriehuys/recipes/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
)

const testDBName = "recipe_test"

func newTestDB(t *testing.T, seedScripts ...string) *pgxpool.Pool {
	ctx := context.Background()

	config, err := config.FromEnvironment()
	if err != nil {
		t.Fatal(err)
	}

	connURL := config.Database.ConnectionURL()
	conn, err := pgx.Connect(ctx, connURL.String())
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if err := conn.Close(ctx); err != nil {
			t.Fatal(err)
		}
	}()

	query := "CREATE DATABASE " + testDBName
	_, err = conn.Exec(ctx, query)
	if err != nil {
		t.Fatal(err)
	}

	if err := conn.Close(ctx); err != nil {
		t.Fatal(err)
	}

	testConnURL := connURL
	testConnURL.Path = testDBName
	pool, err := pgxpool.New(ctx, testConnURL.String())
	if err != nil {
		t.Fatal(err)
	}

	runner := func(conn *pgxpool.Conn) error {
		return runMigrations(ctx, conn.Conn(), migrations.MigrationsFS)
	}
	if err := pool.AcquireFunc(ctx, runner); err != nil {
		pool.Close()
		t.Fatal(err)
	}

	t.Cleanup(func() {
		pool.Close()

		conn, err := pgx.Connect(ctx, connURL.String())
		if err != nil {
			t.Fatal(err)
		}

		defer func() {
			if err := conn.Close(ctx); err != nil {
				t.Fatal(err)
			}
		}()

		query := "DROP DATABASE " + testDBName
		_, err = conn.Exec(ctx, query)
		if err != nil {
			t.Fatal(err)
		}
	})

	for _, scriptFile := range seedScripts {
		script, err := os.ReadFile(scriptFile)
		if err != nil {
			t.Fatal(err)
		}

		_, err = pool.Exec(ctx, string(script))
		if err != nil {
			t.Fatal(err)
		}
	}

	return pool
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
