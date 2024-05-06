package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/cdriehuys/recipes/internal/domain"
)

func loginHandler(logger *slog.Logger, oauthConfig OAuthConfig) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		url := oauthConfig.AuthCodeURL("state")

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})
}

func oauthCallbackHandler(
	logger *slog.Logger,
	oauthConfig OAuthConfig,
	userStore UserStore,
) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

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

		_, err = userStore.RecordLogIn(r.Context(), logger, infoPayload.Id)
		if err != nil {
			logger.Error("Failed to record log in.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// TODO: Only finish registration for new users
		http.Redirect(w, r, "/auth/complete-registration", http.StatusSeeOther)
	})
}

func renderRegistrationForm(w http.ResponseWriter, r *http.Request, templates TemplateWriter, formData, problems map[string]string) error {
	data := map[string]any{
		"formData": formData,
		"problems": problems,
	}

	return templates.Write(w, r, "complete-registration", data)
}

func registerHandler(logger *slog.Logger, templates TemplateWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		if err := renderRegistrationForm(w, r, templates, nil, nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	})
}

func registerFormHandler(logger *slog.Logger, userStore UserStore, templates TemplateWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		userInfo := domain.UserDetails{
			Name: r.FormValue("name"),
		}

		if problems := userInfo.Validate(); len(problems) != 0 {
			logger.Debug("User details failed validation.", "problems", problems)
			formData := map[string]string{"name": userInfo.Name}
			renderRegistrationForm(w, r, templates, formData, problems)
		}

		if err := userStore.UpdateDetails(r.Context(), logger, "?", userInfo); err != nil {
			logger.Error("Failed to update user details.", "error", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	})
}
