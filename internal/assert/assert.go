package assert

import "strings"

type errorer interface {
	Errorf(format string, args ...any)
}

func Equal[T comparable](t errorer, expected, received T) {
	if expected != received {
		t.Errorf("Expected %v, got %v", expected, received)
	}
}

func StringContains(t errorer, haystack, needle string) {
	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected to find %q in %q", needle, haystack)
	}
}
