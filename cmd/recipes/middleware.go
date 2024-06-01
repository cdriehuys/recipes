package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/justinas/nosurf"
)

// noSurf provides CSRF protection for "unsafe" requests.
func (app *application) noSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)
	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.config.Insecure,
	})

	return csrfHandler
}

// requestLogger is a middleware function that logs the request method and URI.
func (app *application) requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.logger.InfoContext(
			r.Context(),
			"Received request.",
			slog.String("method", r.Method),
			slog.String("uri", r.URL.RequestURI()),
		)

		next.ServeHTTP(w, r)
	})
}

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// This header tells Go's HTTP server to close the connection.
				w.Header().Set("Connection", "close")
				app.serverError(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// authenticate adds the ID of an authenticated user to the request context if ID stored in the
// session corresponds to an existing user.
func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := app.sessionManager.GetString(r.Context(), sessionKeyUserID)

		// If there's no ID, the user is unauthenticated and we immediately route to the next
		// handler.
		if id == "" {
			next.ServeHTTP(w, r)
			return
		}

		// Ensure the user actually exists to prevent problems with stale sessions.
		exists, err := app.userModel.Exists(r.Context(), id)
		if err != nil {
			app.serverError(w, r, err)
			return
		}

		// Store the user ID in the request context for access in handlers.
		if exists {
			ctx := context.WithValue(r.Context(), contextKeyUserID, id)
			r = r.WithContext(ctx)
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// If the user is not authenticated, redirect them to the login page and
		// return from the middleware chain so that no subsequent handlers in
		// the chain are executed.
		if !isAuthenticated(r) {
			loginParams := url.Values{}
			loginParams.Set("next", r.URL.Path)

			login := url.URL{
				Path:     "/auth/login",
				RawQuery: loginParams.Encode(),
			}

			http.Redirect(w, r, login.String(), http.StatusSeeOther)
			return
		}

		// Otherwise set the "Cache-Control: no-store" header so that pages
		// require authentication are not stored in the users browser cache (or
		// other intermediary cache).
		w.Header().Add("Cache-Control", "no-store")

		// And call the next handler in the chain.
		next.ServeHTTP(w, r)
	})
}
