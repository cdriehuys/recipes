package main

import (
	"net/http"

	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/justinas/nosurf"
)

type form interface {
	IsValid() bool
}

type templateData struct {
	CSRFToken       string
	IsAuthenticated bool

	Form form

	Recipe  stores.Recipe
	Recipes []stores.Recipe
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CSRFToken:       nosurf.Token(r),
		IsAuthenticated: app.isAuthenticated(r),
	}
}
