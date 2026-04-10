package errs

type Code string

func (c Code) String() string {
	return string(c)
}

const (
	CodeUnauthorized       Code = "unauthorized"
	CodeForbidden          Code = "forbidden"
	CodeInvalid            Code = "invalid"
	CodeUnimplemented      Code = "unimplemented"
	CodeInternal           Code = "internal"
	CodeInternalGenerateID Code = "internalGenerateId"
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

func (e Err) Error() string   { return e.message }
func (e Err) Unwrap() error   { return e.err }
func (e Err) Code() Code      { return e.code }
func (e Err) Message() string { return e.message }

type Unauthorized struct {
	Err
}

func NewUnauthorized() *Unauthorized {
	return &Unauthorized{
		Err: Err{
			message: "unauthorized",
			code:    CodeUnauthorized,
		},
	}
}

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

var Unimplemented = Err{
	message: "unimplemented",
	code:    CodeUnimplemented,
}

type Internal struct {
	Err
}

func NewInternal(message string) *Internal {
	return &Internal{
		Err: Err{
			message: message,
			code:    CodeInternal,
		},
	}
}

func NewInternalErr(message string, err error) *Internal {
	return &Internal{
		Err: Err{
			message: message,
			code:    CodeInternal,
			err:     err,
		},
	}
}

func NewInternalGenerateID(err error) *Internal {
	return &Internal{
		Err: Err{
			message: "failed to generate ID",
			code:    CodeInternalGenerateID,
			err:     err,
		},
	}
}
