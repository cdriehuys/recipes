package mock

import (
	"context"

	"github.com/cdriehuys/recipes/internal/models"
)

type CategoryModel struct {
	LastCreatedCategory models.Category
}

func (model *CategoryModel) Create(_ context.Context, category models.Category) error {
	model.LastCreatedCategory = category

	return nil
}

func (model *CategoryModel) List(context.Context, string) ([]models.Category, error) {
	return nil, nil
}
