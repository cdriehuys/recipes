package mock

import (
	"context"
	"time"

	"github.com/cdriehuys/recipes/internal/models"
	"github.com/google/uuid"
)

var ListedCategories = []models.Category{
	{
		ID:        uuid.New(),
		Name:      "Entrees",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
	{
		ID:        uuid.New(),
		Name:      "Sides",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	},
}

type CategoryModel struct {
	LastCreatedCategory models.Category
}

func (model *CategoryModel) Create(_ context.Context, category models.Category) error {
	model.LastCreatedCategory = category

	return nil
}

func (model *CategoryModel) List(_ context.Context, owner string) ([]models.Category, error) {
	categories := make([]models.Category, len(ListedCategories))
	for index, category := range ListedCategories {
		category.Owner = owner
		categories[index] = category
	}

	return categories, nil
}
