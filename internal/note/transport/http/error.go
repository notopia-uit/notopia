package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/internal/note/domain"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		errorMessage := err.Error()
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, domain.ErrInvalid):
			statusCode = http.StatusBadRequest
		case errors.Is(err, domain.ErrConflict):
			statusCode = http.StatusConflict
		case errors.Is(err, domain.ErrForbidden):
			statusCode = http.StatusForbidden
		case errors.Is(err, domain.ErrUnauthorized):
			statusCode = http.StatusUnauthorized
		case errors.Is(err, domain.ErrInternal):
			statusCode = http.StatusInternalServerError
		default:
		}

		errorResponse := note.Error{
			Code:    "NONE",
			Message: errorMessage,
		}

		c.JSON(statusCode, errorResponse)
	}
}
