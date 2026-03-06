package skillgroup

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict     = errors.New("conflict")
	ErrDependency   = errors.New("dependency violation")
	ErrNotFound     = errors.New("not found")
)
