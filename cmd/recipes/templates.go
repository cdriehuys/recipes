package main

import (
	"net/http"

	"github.com/cdriehuys/recipes/internal/stores"
)

type templateData struct {
	IsAuthenticated bool

	Recipe  stores.Recipe
	Recipes []stores.RecipeListItem
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		IsAuthenticated: app.isAuthenticated(r),
	}
}
