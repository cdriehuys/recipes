package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
	"github.com/cdriehuys/recipes/internal/models/mock"
	"github.com/google/uuid"
)

func Test_application_newRecipe(t *testing.T) {
	app := newTestApp(t)
	server := newTestServer(t, app)

	t.Run("unauthenticated", func(t *testing.T) {
		status, headers, _ := server.get(t, "/new-recipe")

		assert.Equal(t, http.StatusSeeOther, status)
		assertLoginRedirect(t, headers, "/new-recipe")
	})

	t.Run("authenticated", func(t *testing.T) {
		server.authenticate(t, mock.TestUserNormal)

		status, _, _ := server.get(t, "/new-recipe")

		assert.Equal(t, http.StatusOK, status)
	})
}

func Test_application_newRecipePost(t *testing.T) {
	app := newTestApp(t)
	server := newTestServer(t, app)
	server.authenticate(t, mock.TestUserNormal)

	categoryID := uuid.New()

	_, _, formResponse := server.get(t, "/new-recipe")
	csrfToken := extractCSRFToken(t, formResponse)
	t.Logf("Using CSRF token %q", csrfToken)

	testCases := []struct {
		name                  string
		title                 string
		category              string
		instructions          string
		wantStatus            int
		wantValidationMessage string
		wantCreated           bool
	}{
		{
			name:         "valid",
			title:        "Test",
			instructions: "Do the thing.",
			wantStatus:   http.StatusSeeOther,
			wantCreated:  true,
		},
		{
			name:         "valid with category",
			title:        "Test",
			category:     categoryID.String(),
			instructions: "Do the thing.",
			wantStatus:   http.StatusSeeOther,
			wantCreated:  true,
		},
		{
			name:                  "missing title",
			title:                 "",
			instructions:          "Valid.",
			wantStatus:            http.StatusUnprocessableEntity,
			wantValidationMessage: "This field is required.",
		},
		{
			name:                  "title too long",
			title:                 strings.Repeat("a", 201),
			instructions:          "Valid.",
			wantStatus:            http.StatusUnprocessableEntity,
			wantValidationMessage: "This field may not contain more than 200 characters.",
		},
		{
			name:                  "invalid category",
			title:                 "Some title",
			category:              "foo",
			instructions:          "Valid",
			wantStatus:            http.StatusUnprocessableEntity,
			wantValidationMessage: "This field must be a valid category ID.",
		},
		{
			name:                  "missing instructions",
			title:                 "Some Title",
			instructions:          "",
			wantStatus:            http.StatusUnprocessableEntity,
			wantValidationMessage: "This field is required.",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("csrf_token", csrfToken)
			form.Add("title", tt.title)
			form.Add("category", tt.category)
			form.Add("instructions", tt.instructions)

			status, headers, body := server.postForm(t, "/new-recipe", form)

			assert.Equal(t, tt.wantStatus, status)

			if tt.wantValidationMessage != "" {
				assert.StringContains(t, body, tt.wantValidationMessage)
			}

			if tt.wantCreated {
				created := app.recipeModel.(*mock.RecipeModel).LastCreatedRecipe
				assert.Equal(t, mock.TestUserNormal, created.Owner)
				assert.Equal(t, tt.title, created.Title)
				assert.Equal(t, tt.instructions, created.Instructions)

				if tt.category != "" {
					assert.Equal(t, tt.category, created.Category.String())
				}

				assertRedirects(t, headers, "/recipes/"+created.ID.String())
			}
		})
	}
}
