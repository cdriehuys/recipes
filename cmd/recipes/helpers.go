package main

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime/debug"
)

// serverError logs an error-level message including the details of the request that caused the
// error, and returns a generic 500 response to the client.
func (app *application) serverError(w http.ResponseWriter, r *http.Request, err error) {
	var (
		method = r.Method
		uri    = r.URL.RequestURI()
	)

	app.logger.Error(err.Error(), "method", method, "uri", uri)

	if app.config.DevMode {
		trace := debug.Stack()
		fmt.Fprintf(w, "%s\n%s", err.Error(), trace)
		return
	}

	app.clientError(w, http.StatusInternalServerError)
}

// clientError sends back the standard response for the given status code.
func (app *application) clientError(w http.ResponseWriter, status int) {
	http.Error(w, http.StatusText(status), status)
}

// render executes a template and writes it as the response.
func (app *application) render(w http.ResponseWriter, r *http.Request, status int, page string, data templateData) {
	buf := new(bytes.Buffer)
	err := app.templates.Write(buf, r, page, data)
	if err != nil {
		app.serverError(w, r, err)
		return
	}

	// If the template is written to the buffer without any errors, we are safe
	// to write everything to the response. Any errors that occur here are unrecoverable anyways.
	w.WriteHeader(status)
	buf.WriteTo(w)
}

// isAuthenticated returns a boolean indicating if the request session is for an authenticated user.
func (app *application) isAuthenticated(r *http.Request) bool {
	isAuthenticated, ok := r.Context().Value(isAuthenticatedContextKey).(bool)
	if !ok {
		return false
	}

	return isAuthenticated
}
