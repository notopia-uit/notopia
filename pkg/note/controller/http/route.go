package http

import (
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/pkg/common/controller/http"
)

func RegisterRoutes(
	e *gin.Engine,
	httpHandler IHTTPHandler,
) error {
	valdateHandler, err := ValidateHandler()
	if err != nil {
		return err
	}
	// TODO: Implement health later here
	// TODO: Those route with protected from gateway will have user context
	http.RegisterHealthRoutes(e, nil)
	api := e.Group("/")
	{
		api.Use(ErrorHandler())
		api.Use(valdateHandler)
		note.RegisterHandlers(api, httpHandler)
	}
	return nil
}
