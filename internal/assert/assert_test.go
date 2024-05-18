package assert_test

import (
	"fmt"
	"testing"

	"github.com/cdriehuys/recipes/internal/assert"
)

type mockT struct {
	lastErrorf string
}

func (t *mockT) Errorf(format string, args ...any) {
	t.lastErrorf = fmt.Sprintf(format, args...)
}

func TestEqual(t *testing.T) {
	t.Run("Equal values", func(t *testing.T) {
		mockT := mockT{}

		assert.Equal(&mockT, 1, 1)

		if mockT.lastErrorf != "" {
			t.Errorf("Expected no error, got %v", mockT.lastErrorf)
		}
	})

	t.Run("Unequal values", func(t *testing.T) {
		mockT := mockT{}

		assert.Equal(&mockT, 1, 2)

		expected := "Expected 1, got 2"
		if mockT.lastErrorf != expected {
			t.Errorf("Expected %q, got %q", expected, mockT.lastErrorf)
		}
	})
}
