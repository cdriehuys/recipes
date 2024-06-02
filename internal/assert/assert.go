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

// ErrorExists asserts than an error is nil or non-nil based on if it is wanted. If an error is
// wanted and `err` is `nil`, the assertion fails. Similarly, if no error is wanted and `err` is
// non-nil, the assertion fails.
func ErrorExists(t tester, wantErr bool, err error) {
	t.Helper()

	if wantErr {
		if err == nil {
			t.Errorf("Expected error, got nil")
		}
	} else {
		NilError(t, err)
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
