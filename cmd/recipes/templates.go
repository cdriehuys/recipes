package main

import (
	"net/http"

	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/justinas/nosurf"
)

type templateData struct {
	CSRFToken       string
	IsAuthenticated bool

	Recipe  stores.Recipe
	Recipes []stores.RecipeListItem
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CSRFToken:       nosurf.Token(r),
		IsAuthenticated: app.isAuthenticated(r),
	}
}