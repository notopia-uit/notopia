package app

import (
	"errors"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

func hasCode(err error, code errs.Code) bool {
	var e errs.Error
	return errors.As(err, &e) && e.Code() == code
}
