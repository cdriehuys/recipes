package models_test

import (
	"context"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
	"github.com/cdriehuys/recipes/internal/models"
	"github.com/neilotoole/slogt"
)

func Test_Users_Exists(t *testing.T) {
	if testing.Short() {
		t.Skip("models: skipping integration test")
	}

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
			pool := newTestDB(t, "./testdata/seed_users.sql")
			model := models.UserModel{DB: pool, Logger: slogt.New(t)}

			exists, err := model.Exists(context.Background(), tt.id)

			assert.Equal(t, tt.want, exists)
			assert.Equal(t, nil, err)
		})
	}
}
