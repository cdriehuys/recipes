package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/cdriehuys/recipes/internal/domain"
	"github.com/google/uuid"
)

const oauthStateCookie = "recipes.state"

func loginHandler(oauthConfig OAuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

		url := oauthConfig.AuthCodeURL(state.Encode())

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
}

func _oauthNonce(r *http.Request) (string, error) {
	cookie, err := r.Cookie(oauthStateCookie)
	if err != nil {
		return "", err
	}

	return url.QueryUnescape(cookie.Value)
}

func oauthCallbackHandler(
	logger *slog.Logger,
	oauthConfig OAuthConfig,
	session SessionStore,
	userStore UserStore,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		// Remove the nonce cookie immediately.
		http.SetCookie(w, &http.Cookie{
			Name:     oauthStateCookie,
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
		})

		expectedNonce, err := _oauthNonce(r)
		if err != nil {
			logger.Info("No OAuth nonce cookie.", "error", err)

			// TODO: Return template prompting user to retry login flow
			http.Error(w, "Invalid OAuth request.", http.StatusBadRequest)
			return
		}

		rawState := r.URL.Query().Get("state")
		state, err := url.ParseQuery(rawState)
		if err != nil {
			logger.Warn("Malformed state received.", "error", err)
			http.Error(w, "Malformed state parameter.", http.StatusBadRequest)
			return
		}

		receivedNonce := state.Get("nonce")
		if receivedNonce != expectedNonce {
			logger.Warn(
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

		token, err := oauthConfig.Exchange(r.Context(), code)
		if err != nil {
			logger.Error("Failed to convert authorization code to token.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		client := oauthConfig.Client(r.Context(), token)

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
			logger.Error("Failed to decode user info.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		created, err := userStore.RecordLogIn(r.Context(), logger, infoPayload.Id)
		if err != nil {
			logger.Error("Failed to record log in.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if err := session.Create(r.Context(), w, infoPayload.Id); err != nil {
			logger.Error("Failed to create session.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if created {
			http.Redirect(w, r, "/auth/complete-registration", http.StatusSeeOther)
			return
		}

		next, err := url.QueryUnescape(state.Get("next"))
		if err != nil {
			logger.Warn("Malformed next URL.", "url", state.Get("next"))
			next = "/"
		}

		if next == "" {
			next = "/"
		}

		logger.Debug("Redirecting completed OAuth callback.", "next", next)

		http.Redirect(w, r, next, http.StatusSeeOther)
	})
}

func renderRegistrationForm(w http.ResponseWriter, r *http.Request, templates TemplateWriter, formData, problems map[string]string) error {
	data := map[string]any{
		"formData": formData,
		"problems": problems,
	}

	return templates.Write(w, r, "complete-registration", data)
}

func registerHandler(logger *slog.Logger, templates TemplateWriter) AuthenticatedHandler {
	return AuthHandlerFunc(func(w http.ResponseWriter, r *http.Request, userID string) {
		logger := startRequestLogger(r, logger)

		if err := renderRegistrationForm(w, r, templates, nil, nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	})
}

func registerFormHandler(
	logger *slog.Logger,
	userStore UserStore,
	templates TemplateWriter,
) AuthenticatedHandler {
	return AuthHandlerFunc(func(w http.ResponseWriter, r *http.Request, userID string) {
		logger := startRequestLogger(r, logger)

		userInfo := domain.UserDetails{
			Name: r.FormValue("name"),
		}

		if problems := userInfo.Validate(); len(problems) != 0 {
			logger.Debug("User details failed validation.", "problems", problems)
			formData := map[string]string{"name": userInfo.Name}
			renderRegistrationForm(w, r, templates, formData, problems)
			return
		}

		if err := userStore.UpdateDetails(r.Context(), logger, userID, userInfo); err != nil {
			logger.Error("Failed to update user details.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
