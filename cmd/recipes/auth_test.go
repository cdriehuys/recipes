package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

func TestIsAuthenticated(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		name    string
		context context.Context
		want    bool
	}
	testCases := []testCase{
		{
			name:    "missing",
			context: ctx,
			want:    false,
		},
		{
			name:    "wrong type",
			context: context.WithValue(ctx, contextKeyUserID, 3),
			want:    false,
		},
		{
			name:    "empty",
			context: context.WithValue(ctx, contextKeyUserID, ""),
			want:    false,
		},
		{
			name:    "authenticated",
			context: context.WithValue(ctx, contextKeyUserID, "user-id"),
			want:    true,
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(tt.context)

			got := isAuthenticated(req)

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestReqUser(t *testing.T) {
	ctx := context.Background()

	type testCase struct {
		name    string
		context context.Context
		want    string
	}
	testCases := []testCase{
		{
			name:    "missing",
			context: ctx,
			want:    "",
		},
		{
			name:    "wrong type",
			context: context.WithValue(ctx, contextKeyUserID, 3),
			want:    "",
		},
		{
			name:    "empty",
			context: context.WithValue(ctx, contextKeyUserID, ""),
			want:    "",
		},
		{
			name:    "authenticated",
			context: context.WithValue(ctx, contextKeyUserID, "user-id"),
			want:    "user-id",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(tt.context)

			got := reqUser(req)

			assert.Equal(t, tt.want, got)
		})
	}
}
