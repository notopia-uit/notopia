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
	case TypeInternal, TypeUnimplemented:
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
		case TypeUnimplemented:
			code = connect.CodeUnimplemented
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

func FromConnectRPC(err error) *Err {
	if connectErr, ok := errors.AsType[*connect.Error](err); ok {
		var domainErr *Err
		switch connectErr.Code() {
		case connect.CodeInvalidArgument,
			connect.CodeFailedPrecondition,
			connect.CodeOutOfRange:
			domainErr = NewInvalid(connectErr.Error(), "", connectErr)
		case connect.CodeNotFound:
			domainErr = NewNotFound(connectErr.Error(), "", connectErr)
		case connect.CodeAlreadyExists:
			domainErr = NewConflict(connectErr.Error(), "", connectErr)
		case connect.CodePermissionDenied, connect.CodeResourceExhausted:
			domainErr = NewForbidden(connectErr.Error(), "", connectErr)
		case connect.CodeUnauthenticated:
			domainErr = NewUnauthorized(connectErr.Error(), "", connectErr)
		case connect.CodeInternal:
			domainErr = NewInternal(connectErr.Error(), "", connectErr)
		case connect.CodeUnimplemented:
			domainErr = NewUnimplemented()
		case connect.CodeCanceled,
			connect.CodeDeadlineExceeded,
			connect.CodeAborted,
			connect.CodeUnavailable,
			connect.CodeDataLoss,
			connect.CodeUnknown:
			domainErr = NewInternal(connectErr.Error(), "", connectErr)
		}
		return domainErr
	}
	return NewInternal(err.Error(), "", err)
}
