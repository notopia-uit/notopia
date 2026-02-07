package config

import "errors"

var (
	ErrReadFromFile    = errors.New("failed to read configuration from file")
	ErrUnmarshalConfig = errors.New("failed to unmarshal configuration")
	ErrValidateConfig  = errors.New("failed to validate configuration")
)
