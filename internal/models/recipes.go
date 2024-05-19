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

type Recipe struct {
	ID           uuid.UUID `db:"id"`
	Owner        string    `db:"owner"`
	Title        string    `db:"title"`
	Instructions string    `db:"instructions"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

// EditURL returns the URL to the recipe's edit view.
func (r Recipe) EditURL() string {
	return "/recipes/" + r.ID.String() + "/edit"
}

type RecipeModel struct {
	DB     *pgxpool.Pool
	Logger *slog.Logger
}

func (model *RecipeModel) Add(ctx context.Context, recipe Recipe) error {
	query := `
INSERT INTO recipes (id, owner, title, instructions)
VALUES ($1, $2, $3, $4)`

	_, err := model.DB.Exec(
		ctx,
		query,
		recipe.ID,
		recipe.Owner,
		recipe.Title,
		recipe.Instructions,
	)
	if err != nil {
		return fmt.Errorf("failed to insert new recipe: %w", err)
	}

	model.Logger.Info("Persisted new recipe.", "id", recipe.ID)

	return nil
}

func (model *RecipeModel) GetByID(ctx context.Context, owner string, id uuid.UUID) (Recipe, error) {
	query := `SELECT title, instructions, created_at, updated_at
		FROM recipes WHERE owner = $1 AND id = $2`

	recipe := Recipe{ID: id, Owner: owner}
	err := model.DB.QueryRow(ctx, query, owner, id).
		Scan(&recipe.Title, &recipe.Instructions, &recipe.CreatedAt, &recipe.UpdatedAt)
	if err != nil {
		return Recipe{}, fmt.Errorf("failed to query for recipe with ID %s: %w", id, err)
	}

	return recipe, nil
}

func (model *RecipeModel) List(ctx context.Context, owner string) ([]Recipe, error) {
	query := `SELECT id, owner, title, instructions, created_at, updated_at
		FROM recipes WHERE owner = $1 ORDER BY title LIMIT 100`
	rows, err := model.DB.Query(ctx, query, owner)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}
	defer rows.Close()

	recipes, err := pgx.CollectRows(rows, pgx.RowToStructByName[Recipe])
	if err != nil {
		return nil, fmt.Errorf("failed to map recipe rows to struct: %w", err)
	}

	model.Logger.Debug("Retrieved recipe list from database.", "recipes", recipes)

	return recipes, nil
}

func (model *RecipeModel) Update(ctx context.Context, recipe Recipe) error {
	query := `UPDATE recipes
		SET title = $3, instructions = $4
		WHERE owner = $1 AND id = $2`
	result, err := model.DB.Exec(
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
