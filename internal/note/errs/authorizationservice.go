package errs

const (
	CodeAuthorizationServiceInternalError Code = "authorizationServiceInternalError"
)

type errorAuthorizationServiceInternal struct {
	Err
}

func NewAuthorizationInternal(err error) *errorAuthorizationServiceInternal {
	return &errorAuthorizationServiceInternal{
		Err: Err{
			code:    CodeAuthorizationServiceInternalError,
			message: "authorization service internal error",
			err:     err,
		},
	}
}
