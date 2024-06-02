package main

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/cdriehuys/recipes/internal/models"
	"github.com/cdriehuys/recipes/internal/validation"
	"github.com/google/uuid"
)

type RecipeForm struct {
	Category     string
	Title        string
	Instructions string
	validation.Validator
}

func (form *RecipeForm) Validate() {
	form.CheckField(validation.UUIDOrBlank(form.Category), "category", "This field must be a valid category ID.")
	form.CheckField(validation.NotBlank(form.Title), "title", "This field is required.")
	form.CheckField(validation.MaxLength(form.Title, 200), "title", "This field may not contain more than 200 characters.")
	form.CheckField(validation.NotBlank(form.Instructions), "instructions", "This field is required.")
}

func (app *application) addRecipe(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	categories, err := app.categoryModel.List(r.Context(), userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Categories = categories
	data.Form = &RecipeForm{}

	app.render(w, r, http.StatusOK, "add-recipe", data)
}

func (app *application) addRecipePost(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	form := RecipeForm{
		Category:     r.PostFormValue("category"),
		Title:        r.PostFormValue("title"),
		Instructions: r.PostFormValue("instructions"),
	}

	form.Validate()

	if !form.IsValid() {
		app.logger.Debug("New recipe form did not validate.")

		categories, err := app.categoryModel.List(r.Context(), userID)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		data := app.newTemplateData(r)
		data.Categories = categories
		data.Form = &form

		app.render(w, r, http.StatusUnprocessableEntity, "add-recipe", data)
		return
	}

	recipe := models.Recipe{
		ID:           uuid.New(),
		Owner:        userID,
		Title:        form.Title,
		Instructions: form.Instructions,
	}

	if form.Category != "" {
		// We should have validated the ID, so it's okay to panic if it fails to parse.
		categoryID := uuid.MustParse(form.Category)
		recipe.Category = &categoryID
	}

	if err := app.recipeModel.Add(r.Context(), recipe); err != nil {
		app.serverError(w, r, err)
		return
	}

	recipePath, err := url.JoinPath("/recipes", recipe.ID.String())
	if err != nil {
		app.serverError(w, r, fmt.Errorf("failed to build redirect path: %w", err))
		return
	}

	http.Redirect(w, r, recipePath, http.StatusSeeOther)
}
