package stores

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/cdriehuys/recipes/internal/domain"
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

func (s RecipeStore) Add(ctx context.Context, logger *slog.Logger, recipe domain.NewRecipe) error {
	query := `
INSERT INTO recipes (id, owner, title, instructions)
VALUES ($1, $2, $3, $4)`

	if _, err := s.db.Exec(
		ctx,
		query,
		recipe.Id,
		recipe.Owner,
		recipe.Title,
		recipe.Instructions,
	); err != nil {
		return fmt.Errorf("failed to insert new recipe: %w", err)
	}

	logger.Info("Persisted new recipe.", "id", recipe.Id)

	return nil
}

type Recipe struct {
	Title        string
	Instructions string
	CreatedAt    time.Time
}

func (s RecipeStore) GetByID(ctx context.Context, logger *slog.Logger, owner string, id uuid.UUID) (Recipe, error) {
	query := `SELECT title, instructions, created_at FROM recipes WHERE owner = $1 AND id = $2`

	var recipe Recipe
	err := s.db.QueryRow(ctx, query, owner, id).Scan(&recipe.Title, &recipe.Instructions, &recipe.CreatedAt)
	if err != nil {
		return Recipe{}, fmt.Errorf("failed to query for recipe with ID %s: %w", id, err)
	}

	return recipe, nil
}

type RecipeListItem struct {
	Id        uuid.UUID
	Title     string
	CreatedAt time.Time
}

func (s RecipeStore) List(ctx context.Context, logger *slog.Logger, owner string) ([]RecipeListItem, error) {
	query := `SELECT id, title, created_at FROM recipes WHERE owner = $1 LIMIT 100`
	rows, err := s.db.Query(ctx, query, owner)
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
