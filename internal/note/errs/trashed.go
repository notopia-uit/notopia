package errs

import (
	"fmt"
	"time"
)

const (
	CodeTrashedInValid Code = "trashedInValid"
)

type TrashedInValid struct {
	Err
	By string
	At time.Time
}

func NewTrashedInvalid(by string, at time.Time) *TrashedInValid {
	return &TrashedInValid{
		By: by,
		At: at,
		Err: Err{
			message: fmt.Sprintf("invalid trashed value: by=%q, at=%s", by, at.String()),
			code:    CodeTrashedInValid,
		},
	}
}
