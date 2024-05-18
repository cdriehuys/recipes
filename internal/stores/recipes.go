package stores

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Recipe struct {
	ID           uuid.UUID
	Owner        string
	Title        string
	Instructions string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (r Recipe) EditURL() string {
	return "/recipes/" + r.ID.String() + "/edit"
}

type RecipeStore struct {
	db *pgxpool.Pool
}

func NewRecipeStore(db *pgxpool.Pool) RecipeStore {
	return RecipeStore{db}
}

func (s RecipeStore) Add(ctx context.Context, logger *slog.Logger, recipe Recipe) error {
	query := `
INSERT INTO recipes (id, owner, title, instructions)
VALUES ($1, $2, $3, $4)`

	if _, err := s.db.Exec(
		ctx,
		query,
		recipe.ID,
		recipe.Owner,
		recipe.Title,
		recipe.Instructions,
	); err != nil {
		return fmt.Errorf("failed to insert new recipe: %w", err)
	}

	logger.Info("Persisted new recipe.", "id", recipe.ID)

	return nil
}

func (s RecipeStore) GetByID(ctx context.Context, logger *slog.Logger, owner string, id uuid.UUID) (Recipe, error) {
	query := `SELECT title, instructions, created_at, updated_at
		FROM recipes WHERE owner = $1 AND id = $2`

	recipe := Recipe{ID: id, Owner: owner}
	err := s.db.QueryRow(ctx, query, owner, id).
		Scan(&recipe.Title, &recipe.Instructions, &recipe.CreatedAt, &recipe.UpdatedAt)
	if err != nil {
		return Recipe{}, fmt.Errorf("failed to query for recipe with ID %s: %w", id, err)
	}

	return recipe, nil
}

func (s RecipeStore) List(ctx context.Context, logger *slog.Logger, owner string) ([]Recipe, error) {
	query := `SELECT id, owner, title, instructions, created_at, updated_at
		FROM recipes WHERE owner = $1 ORDER BY title LIMIT 100`
	rows, err := s.db.Query(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}
	defer rows.Close()

	recipes, err := pgx.CollectRows(rows, pgx.RowToStructByPos[Recipe])
	if err != nil {
		return nil, fmt.Errorf("failed to map recipe rows to struct: %w", err)
	}

	logger.Debug("Retrieved recipe list from database.", "recipes", recipes)

	return recipes, nil
}

func (s RecipeStore) Update(ctx context.Context, recipe Recipe) error {
	query := `UPDATE recipes
		SET title = $3, instructions = $4
		WHERE owner = $1 AND id = $2`
	result, err := s.db.Exec(
		ctx,
		query,
		recipe.Owner,
		recipe.ID,
		recipe.Title,
		recipe.Instructions,
	)
	if err != nil {
		return fmt.Errorf("failed to update recipe: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
