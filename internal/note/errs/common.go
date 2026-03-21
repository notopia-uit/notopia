package errs

type Code string

var (
	CodeForbidden     Code = "forbidden"
	CodeInvalid       Code = "invalid"
	CodeUnimplemented Code = "unimplemented"
	CodeInternal      Code = "internal"
)

type Error interface {
	error
	Code() Code
}

type Err struct {
	message string
	code    Code
	err     error
}

var _ Error = (*Err)(nil)

func (e Err) Error() string { return e.message }
func (e Err) Unwrap() error { return e.err }
func (e Err) Code() Code    { return e.code }

type Forbidden struct {
	Err
}

func NewForbidden(message string) *Forbidden {
	return &Forbidden{
		Err: Err{
			message: message,
			code:    CodeForbidden,
		},
	}
}

type Invalid struct {
	Err
}

func NewInvalid(message string) *Invalid {
	return &Invalid{
		Err: Err{
			message: message,
			code:    CodeInvalid,
		},
	}
}

type Unimplemented struct {
	Err
}

func NewUnimplemented() *Unimplemented {
	return &Unimplemented{
		Err: Err{
			message: "unimplemented",
			code:    CodeUnimplemented,
		},
	}
}

type Internal struct {
	Err
}

func NewInternal(message string, err error) *Internal {
	return &Internal{
		Err: Err{
			message: message,
			code:    CodeInternal,
			err:     err,
		},
	}
}
