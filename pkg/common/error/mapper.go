package commonerror

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"
)

func ToHTTP(err *Err) (
	message string,
	code string,
	statusCode int,
) {
	message = err.Message()
	code = err.Code()
	switch err.Type() {
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

const (
	ConnectRPCMetadataKeyType    = "x-error-type"
	ConnectRPCMetadataKeyCode    = "x-error-code"
	ConnectRPCMetadataKeyMessage = "x-error-message"
)

func ToConnectRPC(err error) error {
	if err, ok := errors.AsType[*Err](err); ok {
		var code connect.Code

		switch err.Type() {
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

		connectErr := connect.NewError(code, err)

		// Attach metadata to preserve full error information
		connectErr.Meta().Set(ConnectRPCMetadataKeyType, string(err.Type()))
		connectErr.Meta().Set(ConnectRPCMetadataKeyMessage, err.Message())
		if err.Code() != "" {
			connectErr.Meta().Set(ConnectRPCMetadataKeyCode, err.Code())
		}

		return connectErr
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
		// Try to extract metadata first for full error reconstruction
		errorType := connectErr.Meta().Get(ConnectRPCMetadataKeyType)
		errorMessage := connectErr.Meta().Get(ConnectRPCMetadataKeyMessage)
		errorCode := connectErr.Meta().Get(ConnectRPCMetadataKeyCode)

		// If metadata exists, reconstruct the original domain error
		if errorType != "" && errorMessage != "" {
			return New(Type(errorType), errorMessage, errorCode, connectErr)
		}

		// Fallback to code-based mapping if metadata is not available
		var err *Err
		message := connectErr.Error()

		switch connectErr.Code() {
		case connect.CodeInvalidArgument,
			connect.CodeFailedPrecondition,
			connect.CodeOutOfRange:
			err = NewInvalid(message, "", connectErr)
		case connect.CodeNotFound:
			err = NewNotFound(message, "", connectErr)
		case connect.CodeAlreadyExists:
			err = NewConflict(message, "", connectErr)
		case connect.CodePermissionDenied, connect.CodeResourceExhausted:
			err = NewForbidden(message, "", connectErr)
		case connect.CodeUnauthenticated:
			err = NewUnauthorized(message, "", connectErr)
		case connect.CodeInternal:
			err = NewInternal(message, "", connectErr)
		case connect.CodeUnimplemented:
			err = NewUnimplemented()
		case connect.CodeCanceled,
			connect.CodeDeadlineExceeded,
			connect.CodeAborted,
			connect.CodeUnavailable,
			connect.CodeDataLoss,
			connect.CodeUnknown:
			err = NewInternal(message, "", connectErr)
		default:
			err = NewInternal(message, "", connectErr)
		}
		return err
	}
	return NewInternal(err.Error(), "", err)
}
