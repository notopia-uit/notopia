package domain

import "errors"

var (
	ErrConflict = errors.New("resource conflict")
	ErrInternal = errors.New("internal error")
	ErrInvalid  = errors.New("invalid")
	ErrNotFound = errors.New("resource not found")
)
