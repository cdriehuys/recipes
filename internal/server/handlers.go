package server

import (
	"net/http"

	"github.com/google/uuid"
)

func indexHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := state.requestLogger(r)

		if err := state.TemplateEngine.Write(w, "index", nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	}
}

func addRecipeHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := state.requestLogger(r)

		if err := state.TemplateEngine.Write(w, "add-recipe", nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	}
}

func addRecipeFormHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := state.requestLogger(r)

		id := uuid.New()
		title := r.FormValue("title")
		instructions := r.FormValue("instructions")

		query := `INSERT INTO recipes (id, title, instructions) VALUES ($1, $2, $3)`
		if _, err := state.Db.Exec(r.Context(), query, id, title, instructions); err != nil {
			logger.Error("Failed to insert new recipe.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		logger.Info("Inserted new recipe.", "id", id)

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func MakeHandler(state State) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", indexHandler(&state))
	mux.HandleFunc("GET /new-recipe", addRecipeHandler(&state))
	mux.HandleFunc("POST /new-recipe", addRecipeFormHandler(&state))

	return mux
}
