package assert

import "strings"

type tester interface {
	Errorf(format string, args ...any)
	Helper()
}

// Equal fails if the two comparable values are not equivalent.
func Equal[T comparable](t tester, expected, received T) {
	t.Helper()

	if expected != received {
		t.Errorf("Expected %v, got %v", expected, received)
	}
}

// NilError fails if the provided error is not `nil`.
func NilError(t tester, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("Expected nil, got %v", err)
	}
}

// StringContains fails if the substring (needle) is not found in the containing string (haystack).
func StringContains(t tester, haystack, needle string) {
	t.Helper()

	if !strings.Contains(haystack, needle) {
		t.Errorf("Expected to find %q in %q", needle, haystack)
	}
}
