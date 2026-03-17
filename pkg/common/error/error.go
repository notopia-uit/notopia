package commonerror

type Type string

var (
	TypeInternal      Type = "internal error"
	TypeInvalid       Type = "invalid error"
	TypeNotFound      Type = "not found error"
	TypeConflict      Type = "conflict error"
	TypeForbidden     Type = "forbidden error"
	TypeUnauthorized  Type = "unauthorized error"
	TypeUnimplemented Type = "unimplemented error"
)

type Err struct {
	type_   Type
	message string
	code    string
	err     error
}

var _ error = (*Err)(nil)

func New(
	type_ Type,
	message string,
	code string,
	err error,
) *Err {
	return &Err{
		type_:   type_,
		message: message,
		code:    code,
		err:     err,
	}
}

func (d *Err) Type() Type {
	return d.type_
}

func (d *Err) Message() string {
	return d.message
}

func (d *Err) Error() string {
	return d.message
}

func (d *Err) Unwrap() error {
	return d.err
}

func (d *Err) Code() string {
	return d.code
}

func NewInternal(message string, code string, err error) *Err {
	return New(TypeInternal, message, code, err)
}

func NewInvalid(message string, code string, err error) *Err {
	return New(TypeInvalid, message, code, err)
}

func NewNotFound(message string, code string, err error) *Err {
	return New(TypeNotFound, message, code, err)
}

func NewConflict(message string, code string, err error) *Err {
	return New(TypeConflict, message, code, err)
}

func NewForbidden(message string, code string, err error) *Err {
	return New(TypeForbidden, message, code, err)
}

func NewUnauthorized(message string, code string, err error) *Err {
	return New(TypeUnauthorized, message, code, err)
}

func NewUnimplemented() *Err {
	return New(TypeUnimplemented, "Unimplemented", "", nil)
}
