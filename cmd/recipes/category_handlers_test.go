package main

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
	"github.com/cdriehuys/recipes/internal/models/mock"
)

func Test_application_newCategory(t *testing.T) {
	app := newTestApp(t)
	server := newTestServer(t, app)

	t.Run("unauthenticated", func(t *testing.T) {
		status, headers, _ := server.get(t, "/new-category")

		assert.Equal(t, http.StatusSeeOther, status)
		assertLoginRedirect(t, headers, "/new-category")
	})

	t.Run("authenticated", func(t *testing.T) {
		server.authenticate(t, mock.TestUserNormal)

		status, _, _ := server.get(t, "/new-category")
		assert.Equal(t, http.StatusOK, status)
	})
}

func Test_application_newCategoryPost(t *testing.T) {
	app := newTestApp(t)
	server := newTestServer(t, app)
	server.authenticate(t, mock.TestUserNormal)

	_, _, formResponse := server.get(t, "/new-category")
	csrfToken := extractCSRFToken(t, formResponse)
	t.Logf("Using CSRF token %q", csrfToken)

	testCases := []struct {
		name                  string
		category              string
		wantStatus            int
		wantValidationMessage string
		wantCreated           bool
	}{
		{
			name:        "valid",
			category:    "Test",
			wantStatus:  http.StatusSeeOther,
			wantCreated: true,
		},
		{
			name:                  "missing name",
			category:              "",
			wantStatus:            http.StatusUnprocessableEntity,
			wantValidationMessage: "This field is required.",
		},
		{
			name:                  "name too long",
			category:              strings.Repeat("a", 51),
			wantStatus:            http.StatusUnprocessableEntity,
			wantValidationMessage: "This field may not contain more than 50 characters.",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("csrf_token", csrfToken)
			form.Add("name", tt.category)

			status, headers, body := server.postForm(t, "/new-category", form)

			assert.Equal(t, tt.wantStatus, status)

			if tt.wantValidationMessage != "" {
				assert.StringContains(t, body, tt.wantValidationMessage)
			}

			if tt.wantCreated {
				created := app.categoryModel.(*mock.CategoryModel).LastCreatedCategory
				assert.Equal(t, tt.category, created.Name)
				assert.Equal(t, mock.TestUserNormal, created.Owner)

				assertRedirects(t, headers, "/recipes")
			}
		})
	}
}
