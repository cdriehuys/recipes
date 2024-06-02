package assert_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

type mockT struct {
	isHelper   bool
	lastErrorf string
}

func (t *mockT) Errorf(format string, args ...any) {
	t.lastErrorf = fmt.Sprintf(format, args...)
}

func (t *mockT) Helper() {
	t.isHelper = true
}

func TestEqual(t *testing.T) {
	t.Run("Equal values", func(t *testing.T) {
		mockT := mockT{}

		assert.Equal(&mockT, 1, 1)

		if mockT.lastErrorf != "" {
			t.Errorf("Expected no error, got %v", mockT.lastErrorf)
		}

		if mockT.isHelper != true {
			t.Errorf("Expected `t.Helper()` to be called.")
		}
	})

	t.Run("Unequal values", func(t *testing.T) {
		mockT := mockT{}

		assert.Equal(&mockT, 1, 2)

		expected := "Expected 1, got 2"
		if mockT.lastErrorf != expected {
			t.Errorf("Expected %q, got %q", expected, mockT.lastErrorf)
		}

		if mockT.isHelper != true {
			t.Errorf("Expected `t.Helper()` to be called.")
		}
	})
}

func TestNilError(t *testing.T) {
	err := errors.New("some error with a complicated error message")

	testCases := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "error",
			err:  err,
			want: fmt.Sprintf("Expected nil, got %v", err),
		},
		{
			name: "no error",
			err:  nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockT := mockT{}

			assert.NilError(&mockT, tt.err)

			assert.Equal(t, true, mockT.isHelper)
			assert.Equal(t, tt.want, mockT.lastErrorf)
		})
	}
}

func TestStringContains(t *testing.T) {
	testCases := []struct {
		name      string
		needle    string
		haystack  string
		wantError string
	}{
		{
			name:     "string found",
			needle:   "needle",
			haystack: "Does the needle exist in the haystack?",
		},
		{
			name:      "not found",
			needle:    "foo",
			haystack:  "bar",
			wantError: `Expected to find "foo" in "bar"`,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			mockT := mockT{}

			assert.StringContains(&mockT, tt.haystack, tt.needle)

			assert.Equal(t, true, mockT.isHelper)
			assert.Equal(t, tt.wantError, mockT.lastErrorf)
		})
	}
}
