package domain

import "errors"

var (
	ErrNoteNotFound      = errors.New("note not found")
	ErrNoteInvalid       = errors.New("invalid note")
	ErrNoteAlreadyExists = errors.New("note already exists")
)
