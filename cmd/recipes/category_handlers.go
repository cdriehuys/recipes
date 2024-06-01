package main

import (
	"net/http"

	"github.com/cdriehuys/recipes/internal/models"
	"github.com/cdriehuys/recipes/internal/validation"
	"github.com/google/uuid"
)

type categoryForm struct {
	Name string
	validation.Validator
}

func (form *categoryForm) Validate() {
	form.CheckField(validation.NotBlank(form.Name), "name", "This field is required.")
	form.CheckField(validation.MaxLength(form.Name, 50), "name", "This field may not contain more than 50 characters.")
}

func (app *application) newCategory(w http.ResponseWriter, r *http.Request) {
	_ = reqUser(r)

	data := app.newTemplateData(r)
	data.Form = &categoryForm{}

	app.render(w, r, http.StatusOK, "new-category", data)
}

func (app *application) newCategoryPost(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	form := categoryForm{
		Name: r.PostFormValue("name"),
	}
	form.Validate()

	if !form.IsValid() {
		data := app.newTemplateData(r)
		data.Form = &form
		app.render(w, r, http.StatusUnprocessableEntity, "new-category", data)
		return
	}

	category := models.Category{
		ID:    uuid.New(),
		Owner: userID,
		Name:  form.Name,
	}
	if err := app.categoryModel.Create(r.Context(), category); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}
