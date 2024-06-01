package models

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Category struct {
	ID        uuid.UUID `db:"id"`
	Owner     string    `db:"owner"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type CategoryModel struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

func (model *CategoryModel) Create(ctx context.Context, category Category) error {
	query := `INSERT INTO categories (id, owner, name) VALUES ($1, $2, $3)`
	_, err := model.DB.Exec(ctx, query, category.ID, category.Owner, category.Name)
	if err != nil {
		return err
	}

	model.Logger.InfoContext(ctx, "Inserted new category.", "id", category.ID)

	return nil
}

func (model *CategoryModel) List(ctx context.Context, owner string) ([]Category, error) {
	query := `SELECT id, owner, name, created_at, updated_at
		FROM categories
		WHERE owner = $1
		ORDER BY name
		LIMIT 100`
	rows, err := model.DB.Query(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[Category])
	if err != nil {
		return nil, fmt.Errorf("failed to map category rows to struct: %w", err)
	}

	model.Logger.Debug("Retrieved category list from database.", "categories", categories)

	return categories, nil
}
