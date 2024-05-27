package main

import "net/http"

const sessionKeyUserID = "authenticatedUserID"

// isAuthenticated returns a boolean indicating if the request session is for an authenticated user.
func isAuthenticated(r *http.Request) bool {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok {
		return false
	}

	return userID != ""
}

// reqUser returns the ID of the user who made the request. The user is guaranteed to have existed
// when the request was received. This function only works if passed through `app.authenticate`
// first.
func reqUser(r *http.Request) string {
	userID, ok := r.Context().Value(contextKeyUserID).(string)
	if !ok {
		return ""
	}

	return userID
}

// setAuthenticatedUser stores the provided user ID in the session associated with the request as
// the currently authenticated user.
func (app *application) setAuthenticatedUser(r *http.Request, id string) error {
	// Renew the session token to prevent session fixation attacks.
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		return err
	}

	app.sessionManager.Put(r.Context(), sessionKeyUserID, id)

	return nil
}

// clearAuthenticatedUser removes the currently authenticated user's ID from the session so that
// future requests will be unauthenticated.
func (app *application) clearAuthenticatedUser(r *http.Request) error {
	// Renew the session token to prevent session fixation attacks.
	if err := app.sessionManager.RenewToken(r.Context()); err != nil {
		return err
	}

	app.sessionManager.Remove(r.Context(), sessionKeyUserID)

	return nil
}
