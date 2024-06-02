package mock

import (
	"context"

	"github.com/cdriehuys/recipes/internal/models"
	"github.com/google/uuid"
)

type RecipeModel struct {
	LastCreatedRecipe models.Recipe
}

func (model *RecipeModel) Add(_ context.Context, recipe models.Recipe) error {
	model.LastCreatedRecipe = recipe

	return nil
}

func (model *RecipeModel) Delete(context.Context, string, uuid.UUID) error {
	return nil
}

func (model *RecipeModel) GetByID(context.Context, string, uuid.UUID) (models.Recipe, error) {
	return models.Recipe{}, nil
}

func (model *RecipeModel) List(context.Context, string) ([]models.Recipe, error) {
	return nil, nil
}

func (model *RecipeModel) Update(context.Context, models.Recipe) error {
	return nil
}
