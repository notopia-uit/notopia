package http

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
)

func StrictServerErrorHandler(c *gin.Context, err error, statusCode int) {
	message := "internal server error"
	code := ""
	if domainErr, ok := errors.AsType[*commonerror.Err](err); ok {
		message, code, statusCode = commonerror.ToHTTP(domainErr)
	}

	response := note.Error{
		Code:    code,
		Message: message,
	}

	c.JSON(statusCode, response)
}
