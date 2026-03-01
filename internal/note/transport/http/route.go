package http

import (
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
)

func RegisterRoutes(
	e *gin.Engine,
	httpHandler IHTTPHandler,
) error {
	validateHandler, err := ValidateHandler()
	if err != nil {
		return err
	}
	// TODO: Those route with protected from gateway will have user context
	api := e.Group("/")
	{
		api.Use(validateHandler)
		note.RegisterHandlersWithOptions(api, httpHandler, note.GinServerOptions{
			ErrorHandler: StrictServerErrorHandler,
		})
	}
	return nil
}
