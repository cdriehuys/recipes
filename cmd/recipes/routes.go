package main

import (
	"net/http"

	"github.com/cdriehuys/recipes/internal/tracing"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	standard := alice.New(app.recoverPanic, tracing.AddTrace, app.requestLogger)

	mux.Handle(
		"/static/",
		standard.Then(http.StripPrefix("/static/", app.staticServer)),
	)

	dynamic := standard.Append(app.sessionManager.LoadAndSave, app.noSurf, app.authenticate)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.index))
	mux.Handle("GET /auth/callback", dynamic.ThenFunc(app.oauthCallback))
	mux.Handle("GET /auth/login", dynamic.ThenFunc(app.login))
	mux.Handle("GET /privacy-policy", dynamic.ThenFunc(app.privacyPolicy))

	requiresAuth := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /auth/complete-registration", requiresAuth.ThenFunc(app.completeRegistration))
	mux.Handle("POST /auth/complete-registration", requiresAuth.ThenFunc(app.completeRegistrationPost))
	mux.Handle("POST /auth/logout", requiresAuth.ThenFunc(app.logout))
	mux.Handle("GET /new-category", requiresAuth.ThenFunc(app.newCategory))
	mux.Handle("POST /new-category", requiresAuth.ThenFunc(app.newCategoryPost))
	mux.Handle("GET /new-recipe", requiresAuth.ThenFunc(app.addRecipe))
	mux.Handle("POST /new-recipe", requiresAuth.ThenFunc(app.addRecipePost))
	mux.Handle("GET /recipes", requiresAuth.ThenFunc(app.listRecipes))
	mux.Handle("GET /recipes/{recipeID}", requiresAuth.ThenFunc(app.getRecipe))
	mux.Handle("POST /recipes/{recipeID}/delete", requiresAuth.ThenFunc(app.deleteRecipePost))
	mux.Handle("GET /recipes/{recipeID}/edit", requiresAuth.ThenFunc(app.editRecipe))
	mux.Handle("POST /recipes/{recipeID}/edit", requiresAuth.ThenFunc(app.editRecipePost))

	return mux
}
