package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/pkg/api/note"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

func ginMiddlewareErrorHandler(c *gin.Context, message string, statusCode int) {
	var code string
	switch statusCode {
	case 401:
		code = errs.CodeUnauthorized.String()
	case 404:
		code = "pathOrMethodNotFound"
	case 400:
		code = errs.CodeInvalid.String()
	default:
		code = errs.CodeInternal.String()
	}

	response := note.Error{
		Code:     code,
		Message:  message,
		MoreInfo: nil,
	}
	c.JSON(statusCode, response)
}

var _ ginmiddleware.ErrorHandler = ginMiddlewareErrorHandler

func strictServerToHTTPErr(err errs.Error) (
	message string,
	code string,
	statusCode int,
) {
	message = err.Message()
	code = err.Code().String()
	statusCode = 500
	switch err.Code() {
	case errs.CodeUnauthorized:
		statusCode = 401
	case errs.CodeForbidden:
		statusCode = 403
	case errs.CodeInvalid:
		statusCode = 400
	case errs.CodeUnimplemented:
		statusCode = 501
	case errs.CodeInternal,
		errs.CodeInternalGenerateID:
		statusCode = 500

	case errs.CodeAuthorizationServiceInternalError:
		statusCode = 503

	case errs.CodeFolderAlreadyExisted:
		statusCode = 409
	case errs.CodeFolderNotFound:
		statusCode = 404
	case errs.CodeEmptyFolderName:
		statusCode = 400
	case errs.CodeFolderAlreadyTrashed:
		statusCode = 409
	case errs.CodeFoldersNotInWorkspace:
		statusCode = 400
	case errs.CodeDestinationFolderNotInWorkspace:
		statusCode = 400
	case errs.CodeCannotMoveFolderToItOwnSubfolder:
		statusCode = 400

	case errs.CodeNoteNotFound:
		statusCode = 404
	case errs.CodeNoteFailToMarshalDocumentContent:
		statusCode = 500
	case errs.CodeNoteAlreadyTrashed:
		statusCode = 409
	case errs.CodeNotesNotInWorkspace:
		statusCode = 400

	case errs.CodePersistenceInvalid:
		statusCode = 400
	case errs.CodePersistenceInternal:
		statusCode = 500

	case errs.CodeIdentityUserIDInvalid:
		statusCode = 400
	case errs.CodeWorkspaceMembersCannotBeEmpty:
		statusCode = 400
	case errs.CodeWorkspaceMustHaveAtLeastOneOwner:
		statusCode = 400
	case errs.CodeWorkspaceNotFound:
		statusCode = 404
	case errs.CodeWorkspaceBySlugNotFound:
		statusCode = 404
	case errs.CodeWorkspaceRootFolderNotFound:
		statusCode = 404
	case errs.CodeInvalidWorkspaceName:
		statusCode = 400
	case errs.CodeWorkspaceByNoteNotFound:
		statusCode = 404
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

// This is mostly for badRequest handling
func serverErrorHandler(c *gin.Context, err error, statusCode int) {
	if statusCode != http.StatusBadRequest {
		return
	}
	response := note.Error{
		Code:     errs.CodeInvalid.String(),
		Message:  err.Error(),
		MoreInfo: nil,
	}

	c.JSON(statusCode, response)
}

// Register after routing, handling business error
func StrictHandlerErrorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		statusCode := c.Writer.Status()
		// The codegen already ensure setting this appropriately for err
		// Just for ensure
		if statusCode == http.StatusOK || statusCode == 0 {
			statusCode = http.StatusInternalServerError
		}

		// This handle for if we already return strict struct already, then don't
		if c.Writer.Written() {
			return
		}

		message := err.Error()
		code := ""
		if cerr, ok := errors.AsType[errs.Error](err); ok {
			message, code, statusCode = strictServerToHTTPErr(cerr)
		}

		response := note.Error{
			Code:     code,
			Message:  message,
			MoreInfo: nil,
		}

		c.JSON(statusCode, response)
	}
}
