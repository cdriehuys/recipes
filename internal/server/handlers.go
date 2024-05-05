package server

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

type recipeListItem struct {
	Id    uuid.UUID
	Title string
}

func listRecipeHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := state.requestLogger(r)

		query := `SELECT id, title FROM recipes LIMIT 100`
		rows, err := state.Db.Query(r.Context(), query)
		if err != nil {
			logger.Error("Failed to list recipes.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		recipes, err := pgx.CollectRows(rows, pgx.RowToAddrOfStructByPos[recipeListItem])
		if err != nil {
			logger.Error("Failed to read recipes from database.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := state.TemplateEngine.Write(w, "recipe-list", recipes); err != nil {
			logger.Error("Failed to render template.", "error", err)
		}
	}
}

func getRecipeHandler(state *State) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := state.requestLogger(r)

		rawID := r.PathValue("recipeID")
		id, err := uuid.Parse(rawID)
		if err != nil {
			logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		query := `SELECT title, instructions FROM recipes WHERE id = $1`

		var title, instructions string
		if err := state.Db.QueryRow(r.Context(), query, id).Scan(&title, &instructions); err != nil {
			logger.Error("Failed to fetch recipe from database.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		data := map[string]string{
			"title":        title,
			"instructions": instructions,
		}

		if err := state.TemplateEngine.Write(w, "recipe", data); err != nil {
			logger.Error("Failed to render template.", "error", err)
		}
	}
}

func MakeHandler(state State) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /{$}", indexHandler(&state))
	mux.HandleFunc("GET /new-recipe", addRecipeHandler(&state))
	mux.HandleFunc("POST /new-recipe", addRecipeFormHandler(&state))
	mux.HandleFunc("GET /recipes", listRecipeHandler(&state))
	mux.HandleFunc("GET /recipes/{recipeID}", getRecipeHandler(&state))

	return mux
}
