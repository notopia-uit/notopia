package errs

const (
	CodeIdentityUserIDInvalid Code = "identityUserIDInvalid"
)

type errorIdentityUserIDInvalid struct {
	ID string
	Err
}

func NewIdentityUserIDInvalid(id string, err error) *errorIdentityUserIDInvalid {
	return &errorIdentityUserIDInvalid{
		ID: id,
		Err: Err{
			code:    CodeIdentityUserIDInvalid,
			message: "identity user ID is invalid",
			err:     err,
		},
	}
}
