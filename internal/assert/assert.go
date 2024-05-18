package assert

import "testing"

func Equal[T comparable](t *testing.T, expected, received T) {
	if expected != received {
		t.Errorf("Expected %v, got %v", expected, received)
	}
}
