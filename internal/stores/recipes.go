package stores

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RecipeStore struct {
	db *pgxpool.Pool
}

func NewRecipeStore(db *pgxpool.Pool) RecipeStore {
	return RecipeStore{db}
}

type NewRecipe struct {
	Id           uuid.UUID
	Title        string
	Instructions string
}

func (s RecipeStore) Add(ctx context.Context, logger *slog.Logger, recipe NewRecipe) error {
	query := `
INSERT INTO recipes (id, title, instructions)
VALUES ($1, $2, $3)`

	if _, err := s.db.Exec(ctx, query, recipe.Id, recipe.Title, recipe.Instructions); err != nil {
		return fmt.Errorf("failed to insert new recipe: %w", err)
	}

	logger.Info("Persisted new recipe.", "id", recipe.Id)

	return nil
}

type Recipe struct {
	Title        string
	Instructions string
}

func (s RecipeStore) GetByID(ctx context.Context, logger *slog.Logger, id uuid.UUID) (Recipe, error) {
	query := `SELECT title, instructions FROM recipes WHERE id = $1`

	var recipe Recipe
	err := s.db.QueryRow(ctx, query, id).Scan(&recipe.Title, &recipe.Instructions)
	if err != nil {
		return Recipe{}, fmt.Errorf("failed to query for recipe with ID %s: %w", id, err)
	}

	return recipe, nil
}

type RecipeListItem struct {
	Id    uuid.UUID
	Title string
}

func (s RecipeStore) List(ctx context.Context, logger *slog.Logger) ([]RecipeListItem, error) {
	query := `SELECT id, title FROM recipes LIMIT 100`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}
	defer rows.Close()

	recipes, err := pgx.CollectRows(rows, pgx.RowToStructByPos[RecipeListItem])
	if err != nil {
		return nil, fmt.Errorf("failed to map recipe rows to struct: %w", err)
	}

	logger.Debug("Retrieved recipe list from database.", "recipes", recipes)

	return recipes, nil
}
