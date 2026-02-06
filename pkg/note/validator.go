package note

import (
	"github.com/go-playground/validator/v10"
)

func ProvideValidate() *validator.Validate {
	return validator.New(
		validator.WithRequiredStructEnabled(),
	)
}
