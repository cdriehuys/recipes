package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"

	"github.com/cdriehuys/recipes/internal/models"
	"github.com/cdriehuys/recipes/internal/validation"
	"github.com/google/uuid"
)

func (app *application) index(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "index", app.newTemplateData(r))
}

func (app *application) privacyPolicy(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "privacy-policy", app.newTemplateData(r))
}

func (app *application) oauthCallback(w http.ResponseWriter, r *http.Request) {
	expectedNonce := app.sessionManager.PopString(r.Context(), "oauth-nonce")
	if expectedNonce == "" {
		app.logger.Info("No OAuth nonce present in session.")

		// TODO: Return template prompting user to retry login flow
		http.Error(w, "Invalid OAuth request.", http.StatusBadRequest)
		return
	}

	rawState := r.URL.Query().Get("state")
	state, err := url.ParseQuery(rawState)
	if err != nil {
		app.logger.Warn("Malformed state received.", "error", err)
		http.Error(w, "Malformed state parameter.", http.StatusBadRequest)
		return
	}

	receivedNonce := state.Get("nonce")
	if receivedNonce != expectedNonce {
		app.logger.Warn(
			"Mismatched nonce. Possibly tampered OAuth flow.",
			"expected",
			expectedNonce,
			"received",
			receivedNonce,
		)

		// TODO: Template response
		http.Error(w, "Invalid OAuth request.", http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")

	token, err := app.oauthConfig.Exchange(r.Context(), code)
	if err != nil {
		app.logger.Error("Failed to convert authorization code to token.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	client := app.oauthConfig.Client(r.Context(), token)

	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	var infoPayload struct {
		Id string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&infoPayload); err != nil {
		app.logger.Error("Failed to decode user info.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	created, err := app.userModel.RecordLogIn(r.Context(), infoPayload.Id)
	if err != nil {
		app.logger.Error("Failed to record log in.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.setAuthenticatedUser(r, infoPayload.Id); err != nil {
		app.serverError(w, r, err)
		return
	}

	if created {
		http.Redirect(w, r, "/auth/complete-registration", http.StatusSeeOther)
		return
	}

	next, err := url.QueryUnescape(state.Get("next"))
	if err != nil {
		app.logger.Warn("Malformed next URL.", "url", state.Get("next"))
		next = "/"
	}

	if next == "" {
		next = "/"
	}

	app.logger.Debug("Redirecting completed OAuth callback.", "next", next)

	http.Redirect(w, r, next, http.StatusSeeOther)
}

type registrationForm struct {
	Name string
	validation.Validator
}

func (form *registrationForm) Validate() {
	form.CheckField(validation.NotBlank(form.Name), "name", "This field may not be blank.")
	form.CheckField(validation.MaxLength(form.Name, 50), "name", "This field may not contain more than 50 characters.")
}

func (app *application) completeRegistration(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = &registrationForm{}

	app.render(w, r, http.StatusOK, "complete-registration", data)
}

func (app *application) completeRegistrationPost(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	form := registrationForm{
		Name: r.PostFormValue("name"),
	}
	form.Validate()

	if !form.IsValid() {
		app.logger.Debug("User details failed validation.")
		data := app.newTemplateData(r)
		data.Form = &form
		app.render(w, r, http.StatusOK, "complete-registration", data)
		return
	}

	if err := app.userModel.UpdateName(r.Context(), userID, form.Name); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	nonce := uuid.New().String()
	app.sessionManager.Put(r.Context(), "oauth-nonce", nonce)

	state := url.Values{}
	state.Set("next", r.URL.Query().Get("next"))
	state.Set("nonce", nonce)

	url := app.oauthConfig.AuthCodeURL(state.Encode())

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	if err := app.clearAuthenticatedUser(r); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) listRecipes(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	recipes, err := app.recipeModel.List(r.Context(), userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Recipes = recipes

	app.render(w, r, http.StatusOK, "recipe-list", data)
}

func (app *application) getRecipe(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	rawID := r.PathValue("recipeID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		app.logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	recipe, err := app.recipeModel.GetByID(r.Context(), userID, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Recipe = recipe

	app.render(w, r, http.StatusOK, "recipe", data)
}

func (app *application) editRecipe(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	rawID := r.PathValue("recipeID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		app.logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	recipe, err := app.recipeModel.GetByID(r.Context(), userID, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	form := RecipeForm{
		Title:        recipe.Title,
		Instructions: recipe.Instructions,
	}

	data := app.newTemplateData(r)
	data.Form = &form
	data.Recipe = recipe

	app.render(w, r, http.StatusOK, "edit-recipe", data)
}

func (app *application) editRecipePost(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	rawID := r.PathValue("recipeID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		app.logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
		app.clientError(w, http.StatusNotFound)
		return
	}

	form := RecipeForm{
		Title:        r.PostFormValue("title"),
		Instructions: r.PostFormValue("instructions"),
	}
	form.Validate()

	if !form.IsValid() {
		recipe, err := app.recipeModel.GetByID(r.Context(), userID, id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		data := app.newTemplateData(r)
		data.Form = &form
		data.Recipe = recipe

		app.render(w, r, http.StatusOK, "edit-recipe", data)
		return
	}

	recipe := models.Recipe{
		ID:           id,
		Owner:        userID,
		Title:        form.Title,
		Instructions: form.Instructions,
	}
	if err := app.recipeModel.Update(r.Context(), recipe); err != nil {
		if errors.Is(err, models.ErrNotFound) {
			app.clientError(w, http.StatusNotFound)
		} else {
			app.serverError(w, r, err)
		}
		return
	}

	recipeURL, err := url.JoinPath("/recipes", id.String())
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, recipeURL, http.StatusSeeOther)
}

func (app *application) deleteRecipePost(w http.ResponseWriter, r *http.Request) {
	userID := reqUser(r)

	rawID := r.PathValue("recipeID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		app.logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
		app.clientError(w, http.StatusNotFound)
		return
	}

	if err := app.recipeModel.Delete(r.Context(), userID, id); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}
