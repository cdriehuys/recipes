package models_test

import (
	"context"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
	"github.com/cdriehuys/recipes/internal/models"
	"github.com/neilotoole/slogt"
)

func newUserModel(t *testing.T) *models.UserModel {
	pool := newTestDB(t, "./testdata/seed_users.sql")

	return &models.UserModel{DB: pool, Logger: slogt.New(t)}
}

func Test_UserModel_Exists(t *testing.T) {
	markAsIntegrationTest(t)

	testCases := []struct {
		name string
		id   string
		want bool
	}{
		{
			name: "Valid ID",
			id:   "1",
			want: true,
		},
		{
			name: "Missing ID",
			id:   "2",
			want: false,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			model := newUserModel(t)

			exists, err := model.Exists(context.Background(), tt.id)

			assert.Equal(t, tt.want, exists)
			assert.NilError(t, err)
		})
	}
}

func Test_UserModel_RecordLogIn(t *testing.T) {
	markAsIntegrationTest(t)

	testCases := []struct {
		name        string
		id          string
		wantCreated bool
	}{
		{
			name:        "existing user",
			id:          "1",
			wantCreated: false,
		},
		{
			name:        "new user",
			id:          "2",
			wantCreated: true,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			model := newUserModel(t)

			created, err := model.RecordLogIn(context.Background(), tt.id)

			assert.Equal(t, tt.wantCreated, created)
			assert.NilError(t, err)
		})
	}
}

func Test_UserModel_UpdateName(t *testing.T) {
	markAsIntegrationTest(t)

	testCases := []struct {
		name     string
		id       string
		userName string
	}{
		{
			name:     "Valid ID",
			id:       "1",
			userName: "John Doe",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			model := newUserModel(t)

			err := model.UpdateName(context.Background(), tt.id, tt.userName)

			assert.NilError(t, err)
		})
	}
}
