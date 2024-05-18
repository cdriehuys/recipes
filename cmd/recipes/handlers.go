package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/cdriehuys/recipes/internal/domain"
	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/cdriehuys/recipes/internal/validation"
	"github.com/google/uuid"
)

const oauthStateCookie = "recipes.state"

func (app *application) index(w http.ResponseWriter, r *http.Request) {
	app.render(w, r, http.StatusOK, "index", app.newTemplateData(r))
}

func _oauthNonce(r *http.Request) (string, error) {
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		return "", err
	}

	return url.QueryUnescape(cookie.Value)
}

func (app *application) oauthCallback(w http.ResponseWriter, r *http.Request) {
	// Remove the nonce cookie immediately.
	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookie,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})

	expectedNonce, err := _oauthNonce(r)
	if err != nil {
		app.logger.Info("No OAuth nonce cookie.", "error", err)

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

	created, err := app.userStore.RecordLogIn(r.Context(), app.logger, infoPayload.Id)
	if err != nil {
		app.logger.Error("Failed to record log in.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.logger.Error("Failed to renew session token.", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	app.sessionManager.Put(r.Context(), "authenticatedUserID", infoPayload.Id)

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

func renderRegistrationForm(w http.ResponseWriter, r *http.Request, templates templateWriter, formData, problems map[string]string) error {
	data := map[string]any{
		"formData": formData,
		"problems": problems,
	}

	return templates.Write(w, r, "complete-registration", data)
}

func (app *application) completeRegistration(w http.ResponseWriter, r *http.Request) {
	if err := renderRegistrationForm(w, r, app.templates, nil, nil); err != nil {
		app.logger.Error("Failed to execute template.", "error", err)
	}
}

func (app *application) completeRegistrationPost(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")
	if userID == "" {
		app.logger.Error("Unauthenticated user made it to registration completion.")
		app.clientError(w, http.StatusInternalServerError)
		return
	}

	userInfo := domain.UserDetails{
		Name: r.FormValue("name"),
	}

	if problems := userInfo.Validate(); len(problems) != 0 {
		app.logger.Debug("User details failed validation.", "problems", problems)
		formData := map[string]string{"name": userInfo.Name}
		renderRegistrationForm(w, r, app.templates, formData, problems)
		return
	}

	if err := app.userStore.UpdateDetails(r.Context(), app.logger, userID, userInfo); err != nil {
		app.serverError(w, r, err)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {
	nonce := uuid.New().String()
	cookie := http.Cookie{
		Name:     oauthStateCookie,
		Value:    url.QueryEscape(nonce),
		MaxAge:   int((5 * time.Minute).Seconds()),
		HttpOnly: true,

		// Because the OAuth callback is a redirect from a different site, this cannot be set to
		// `Strict`.
		SameSite: http.SameSiteLaxMode,
	}
	http.SetCookie(w, &cookie)

	state := url.Values{}
	state.Set("next", r.URL.Query().Get("next"))
	state.Set("nonce", nonce)

	url := app.oauthConfig.AuthCodeURL(state.Encode())

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (app *application) logout(w http.ResponseWriter, r *http.Request) {
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		app.serverError(w, r, err)
		return
	}

	app.sessionManager.Remove(r.Context(), "authenticatedUserID")

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type RecipeForm struct {
	Title        string
	Instructions string
	validation.Validator
}

func (form *RecipeForm) Validate() {
	form.CheckField(validation.NotBlank(form.Title), "title", "This field is required.")
	form.CheckField(validation.MaxLength(form.Title, 200), "title", "This field may not contain more than 200 characters.")
	form.CheckField(validation.NotBlank(form.Instructions), "instructions", "This field is required.")
}

func (app *application) addRecipe(w http.ResponseWriter, r *http.Request) {
	data := app.newTemplateData(r)
	data.Form = &RecipeForm{}

	app.render(w, r, http.StatusOK, "add-recipe", data)
}

func (app *application) addRecipePost(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	form := RecipeForm{
		Title:        r.PostFormValue("title"),
		Instructions: r.PostFormValue("instructions"),
	}

	form.Validate()

	if !form.IsValid() {
		app.logger.Debug("New recipe form did not validate.")

		data := app.newTemplateData(r)
		data.Form = &form

		app.render(w, r, http.StatusOK, "add-recipe", data)
		return
	}

	recipe := stores.Recipe{
		ID:           uuid.New(),
		Owner:        userID,
		Title:        form.Title,
		Instructions: form.Instructions,
	}

	if err := app.recipeStore.Add(r.Context(), app.logger, recipe); err != nil {
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

func (app *application) listRecipes(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	recipes, err := app.recipeStore.List(r.Context(), app.logger, userID)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Recipes = recipes

	app.render(w, r, http.StatusOK, "recipe-list", data)
}

func (app *application) getRecipe(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	rawID := r.PathValue("recipeID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		app.logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	recipe, err := app.recipeStore.GetByID(r.Context(), app.logger, userID, id)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	data := app.newTemplateData(r)
	data.Recipe = recipe

	app.render(w, r, http.StatusOK, "recipe", data)
}

func (app *application) editRecipe(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

	rawID := r.PathValue("recipeID")
	id, err := uuid.Parse(rawID)
	if err != nil {
		app.logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	recipe, err := app.recipeStore.GetByID(r.Context(), app.logger, userID, id)
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

	app.render(w, r, http.StatusOK, "edit-recipe", data)
}

func (app *application) editRecipePost(w http.ResponseWriter, r *http.Request) {
	userID := app.sessionManager.GetString(r.Context(), "authenticatedUserID")

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
		data := app.newTemplateData(r)
		data.Form = &form
		app.render(w, r, http.StatusOK, "edit-recipe", data)
		return
	}

	recipe := stores.Recipe{
		ID:           id,
		Owner:        userID,
		Title:        form.Title,
		Instructions: form.Instructions,
	}
	if err := app.recipeStore.Update(r.Context(), recipe); err != nil {
		if errors.Is(err, stores.ErrNotFound) {
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
