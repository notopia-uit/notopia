package domain

import (
	"errors"
)

var (
	ErrInternal     = errors.New("internal error")
	ErrInvalid      = errors.New("invalid data")
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
)
