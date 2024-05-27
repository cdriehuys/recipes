package main

import (
	"net/http"

	"github.com/cdriehuys/recipes/internal/models"
	"github.com/justinas/nosurf"
)

type form interface {
	IsValid() bool
}

type templateData struct {
	CSRFToken       string
	IsAuthenticated bool

	Form form

	Recipe  models.Recipe
	Recipes []models.Recipe
}

func (app *application) newTemplateData(r *http.Request) templateData {
	return templateData{
		CSRFToken:       nosurf.Token(r),
		IsAuthenticated: isAuthenticated(r),
	}
}
