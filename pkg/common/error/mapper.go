package commonerror

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
)

func ToHTTP(domainErr *Err) (
	message string,
	code string,
	statusCode int,
) {
	message = domainErr.Message()
	code = domainErr.Code()
	switch domainErr.Type() {
	case TypeInvalid:
		statusCode = http.StatusBadRequest
	case TypeNotFound:
		statusCode = http.StatusNotFound
	case TypeConflict:
		statusCode = http.StatusConflict
	case TypeForbidden:
		statusCode = http.StatusForbidden
	case TypeUnauthorized:
		statusCode = http.StatusUnauthorized
	case TypeInternal:
		statusCode = http.StatusInternalServerError
	}
	return
}

func ToConnectRPC(err error) error {
	if domainErr, ok := errors.AsType[*Err](err); ok {
		var code connect.Code

		switch domainErr.Type() {
		case TypeInvalid:
			code = connect.CodeInvalidArgument
		case TypeNotFound:
			code = connect.CodeNotFound
		case TypeConflict:
			code = connect.CodeAlreadyExists
		case TypeForbidden:
			code = connect.CodePermissionDenied
		case TypeUnauthorized:
			code = connect.CodeUnauthenticated
		case TypeInternal:
			code = connect.CodeInternal
		default:
			code = connect.CodeUnknown
		}
		return connect.NewError(code, err)
	}

	if errors.Is(err, context.Canceled) {
		return connect.NewError(connect.CodeCanceled, err)
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return connect.NewError(connect.CodeDeadlineExceeded, err)
	}

	return connect.NewError(connect.CodeInternal, err)
}
