package stores

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/cdriehuys/recipes/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserStore struct {
	db *pgxpool.Pool
}

func NewUserStore(db *pgxpool.Pool) UserStore {
	return UserStore{db}
}

// Record the log in for a user. Returns a boolean indicating if the user needs to complete their
// registration as well as any error that occurred.
func (s UserStore) RecordLogIn(ctx context.Context, logger *slog.Logger, id string) (bool, error) {
	// This process is done as 2 separate queries to make it easy to tell if we added a new user.
	// With one query that upserted a user and set the new login time, it was difficult to tell if
	// the user was already present in the database.

	insert := `INSERT INTO "users" (id, name) VALUES ($1, '') ON CONFLICT DO NOTHING`
	_, err := s.db.Exec(ctx, insert, id)
	if err != nil {
		return false, fmt.Errorf("failed to persist user: %w", err)
	}

	logger.Info("Authenticated user persisted.", "id", id)

	loginUpdate := `UPDATE "users" SET last_login = now() WHERE id = $1 RETURNING name`
	var name string
	if err := s.db.QueryRow(ctx, loginUpdate, id).Scan(&name); err != nil {
		return false, fmt.Errorf("failed to update user login time: %w", err)
	}

	// If the user has no name, that's a signal they need to complete their registration.
	return name == "", nil
}

func (s UserStore) UpdateDetails(ctx context.Context, logger *slog.Logger, id string, userDetails domain.UserDetails) error {
	query := `UPDATE "users" SET name = $2 WHERE id = $1`
	if _, err := s.db.Exec(ctx, query, id, userDetails.Name); err != nil {
		return fmt.Errorf("failed to update user details: %w", err)
	}

	logger.Info("Updated user details.", "id", id)

	return nil
}
