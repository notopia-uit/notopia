package http

import (
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/api/note"
	"github.com/oapi-codegen/gin-middleware"
)

func ValidateHandler() (gin.HandlerFunc, error) {
	spec, err := note.GetOpenAPI(nil)
	if err != nil {
		return nil, err
	}
	spec.Servers = nil
	return ginmiddleware.OapiRequestValidator(spec), nil
}
