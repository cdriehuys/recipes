package models

import "errors"

// ErrNotFound indicates the object targeted by the query could not be found.
var ErrNotFound = errors.New("object not found")
