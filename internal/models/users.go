package models

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID        string             `db:"id"`
	Name      string             `db:"name"`
	CreatedAt time.Time          `db:"created_at"`
	UpdatedAt time.Time          `db:"updated_at"`
	LastLogin pgtype.Timestamptz `db:"last_login"`
}

type UserModel struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

func (model *UserModel) Exists(ctx context.Context, id string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM "users" WHERE id = $1)`

	var exists bool
	err := model.DB.QueryRow(ctx, query, id).Scan(&exists)

	return exists, err
}

// Record the log in for a user. Returns a boolean indicating if the user needs to complete their
// registration as well as any error that occurred.
func (model *UserModel) RecordLogIn(ctx context.Context, id string) (bool, error) {
	// This process is done as 2 separate queries to make it easy to tell if we added a new user.
	// With one query that upserted a user and set the new login time, it was difficult to tell if
	// the user was already present in the database.

	insert := `INSERT INTO "users" (id, name) VALUES ($1, '') ON CONFLICT DO NOTHING`
	_, err := model.DB.Exec(ctx, insert, id)
	if err != nil {
		return false, fmt.Errorf("failed to persist user: %w", err)
	}

	model.Logger.Info("Authenticated user persisted.", "id", id)

	loginUpdate := `UPDATE "users" SET last_login = now() WHERE id = $1 RETURNING name`
	var name string
	if err := model.DB.QueryRow(ctx, loginUpdate, id).Scan(&name); err != nil {
		return false, fmt.Errorf("failed to update user login time: %w", err)
	}

	// If the user has no name, that's a signal they need to complete their registration.
	return name == "", nil
}

func (model *UserModel) UpdateName(ctx context.Context, id, name string) error {
	query := `UPDATE "users" SET name = $2 WHERE id = $1`
	if _, err := model.DB.Exec(ctx, query, id, name); err != nil {
		return fmt.Errorf("failed to update user details: %w", err)
	}

	model.Logger.Info("Updated user details.", "id", id)

	return nil
}
