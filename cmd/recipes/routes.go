package main

import (
	"net/http"

	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	standard := alice.New(app.recoverPanic, app.requestLogger)

	mux.Handle(
		"/static/",
		standard.Then(http.StripPrefix("/static/", app.staticServer)),
	)

	dynamic := standard.Append(app.sessionManager.LoadAndSave, app.authenticate)

	mux.Handle("GET /{$}", dynamic.ThenFunc(app.index))
	mux.Handle("GET /auth/callback", dynamic.ThenFunc(app.oauthCallback))
	mux.Handle("GET /auth/login", dynamic.ThenFunc(app.login))

	requiresAuth := dynamic.Append(app.requireAuthentication)

	mux.Handle("GET /auth/complete-registration", requiresAuth.ThenFunc(app.completeRegistration))
	mux.Handle("POST /auth/complete-registration", requiresAuth.ThenFunc(app.completeRegistrationPost))
	mux.Handle("GET /new-recipe", requiresAuth.ThenFunc(app.addRecipe))
	mux.Handle("POST /new-recipe", requiresAuth.ThenFunc(app.addRecipePost))
	mux.Handle("GET /recipes", requiresAuth.ThenFunc(app.listRecipes))
	mux.Handle("GET /recipes/{recipeID}", requiresAuth.ThenFunc(app.getRecipe))

	return mux
}
