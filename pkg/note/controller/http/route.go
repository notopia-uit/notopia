package http

import (
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/pkg/common/controller/http"
)

func RegisterRoutes(
	e *gin.Engine,
	httpHandler IHTTPHandler,
	healthManager *http.HealthManager,
) error {
	valdateHandler, err := ValidateHandler()
	if err != nil {
		return err
	}
	http.RegisterHealthRoutes(e, healthManager)
	// TODO: Those route with protected from gateway will have user context
	api := e.Group("/")
	{
		api.Use(ErrorHandler())
		api.Use(valdateHandler)
		note.RegisterHandlers(api, httpHandler)
	}
	return nil
}
