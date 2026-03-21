package errs

var CodeAuthorizationUnavailable Code = "authorization_1"

type ErrorAuthorizationUnavailable struct {
	Err
}

func NewErrorAuthorizationUnavailable(err error) *ErrorAuthorizationUnavailable {
	return &ErrorAuthorizationUnavailable{
		Err: Err{
			code:    CodeAuthorizationUnavailable,
			message: "authorization service is unavailable",
			err:     err,
		},
	}
}
