package routes

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/cdriehuys/recipes/internal/stores"
	"github.com/google/uuid"
)

func addRecipeHandler(logger *slog.Logger, templates TemplateWriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		if err := templates.Write(w, "add-recipe", nil); err != nil {
			logger.Error("Failed to execute template.", "error", err)
		}
	})
}

func addRecipeFormHandler(
	logger *slog.Logger,
	recipeStore RecipeStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		recipe := stores.NewRecipe{
			Id:           uuid.New(),
			Title:        r.FormValue("title"),
			Instructions: r.FormValue("instructions"),
		}

		if err := recipeStore.Add(r.Context(), logger, recipe); err != nil {
			logger.Error("Failed to save new recipe.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		recipePath, err := url.JoinPath("/recipes", recipe.Id.String())
		if err != nil {
			logger.Error("Failed to build redirect path.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, recipePath, http.StatusSeeOther)
	}
}

func listRecipeHandler(
	logger *slog.Logger,
	recipeStore RecipeStore,
	templates TemplateWriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		recipes, err := recipeStore.List(r.Context(), logger)
		if err != nil {
			logger.Error("Failed to read recipes from database.", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := templates.Write(w, "recipe-list", recipes); err != nil {
			logger.Error("Failed to render template.", "error", err)
		}
	}
}

func getRecipeHandler(
	logger *slog.Logger,
	recipeStore RecipeStore,
	templates TemplateWriter,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := startRequestLogger(r, logger)

		rawID := r.PathValue("recipeID")
		id, err := uuid.Parse(rawID)
		if err != nil {
			logger.Debug("Received invalid recipe ID", "id", rawID, "error", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		recipe, err := recipeStore.GetByID(r.Context(), logger, id)
		if err != nil {
			logger.Error("Failed to fetch recipe by ID.", "id", id, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if err := templates.Write(w, "recipe", recipe); err != nil {
			logger.Error("Failed to render template.", "error", err)
		}
	}
}
