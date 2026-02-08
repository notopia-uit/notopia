package domain

import "errors"

var (
	ErrConflict     = errors.New("resource conflict")
	ErrInternal     = errors.New("internal error")
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("resource not found")
)
