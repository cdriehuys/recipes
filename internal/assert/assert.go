package assert

type errorer interface {
	Errorf(format string, args ...any)
}

func Equal[T comparable](t errorer, expected, received T) {
	if expected != received {
		t.Errorf("Expected %v, got %v", expected, received)
	}
}
