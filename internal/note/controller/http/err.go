package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func toHTTPErr(err *errs.Err) (
	message string,
	code string,
	statusCode int,
) {
	message = err.Message()
	code = err.Code().String()
	statusCode = 500
	switch err.Code() {
	case errs.CodeForbidden:
		statusCode = 403
	case errs.CodeInvalid:
		statusCode = 400
	case errs.CodeUnimplemented:
		statusCode = 501
	case errs.CodeInternal:
		statusCode = 500

	case errs.CodeAuthorizationServiceInternalError:
		statusCode = 503

	case errs.CodeFolderNotFound:
		statusCode = 404
	case errs.CodeEmptyFolderName:
		statusCode = 400
	case errs.CodeFolderAlreadyTrashed:
		statusCode = 409

	case errs.CodeNoteNotFound:
		statusCode = 404
	case errs.CodeNoteFailToMarshalDocumentContent:
		statusCode = 500
	case errs.CodeNoteAlreadyTrashed:
		statusCode = 409

	case errs.CodePersistenceInvalid:
		statusCode = 400
	case errs.CodePersistenceInternal:
		statusCode = 500

	case errs.CodeWorkspaceNotFound:
		statusCode = 404
	case errs.CodeWorkspaceBySlugNotFound:
		statusCode = 404
	case errs.CodeWorkspaceRootFolderNotFound:
		statusCode = 404
	case errs.CodeInvalidWorkspaceName:
		statusCode = 400
	case errs.CodeInvalidWorkspaceSlug:
		statusCode = 400
	case errs.CodeWorkspaceSlugAlreadyExists:
		statusCode = 409

	case errs.CodeWorkspaceEventPubSubFailedToCreateMessage:
		statusCode = 500
	case errs.CodeWorkspaceEventPubSubPublishFailed:
		statusCode = 500
	case errs.CodeWorkspaceEventPubSubSubscribeFailed:
		statusCode = 500
	}
	return
}

func strictServerErrorHandler(c *gin.Context, err error, statusCode int) {
	message := err.Error()
	code := ""
	if cerr, ok := errors.AsType[*errs.Err](err); ok {
		message, code, statusCode = toHTTPErr(cerr)
	}

	response := note.Error{
		Code:     code,
		Message:  message,
		MoreInfo: nil,
	}

	c.JSON(statusCode, response)
}
