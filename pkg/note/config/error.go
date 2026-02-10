package config

import "errors"

var (
	ErrReadFromFile = errors.New("failed to read configuration from file")
	ErrUnmarshal    = errors.New("failed to unmarshal configuration")
	ErrValidate     = errors.New("failed to validate configuration")
)
