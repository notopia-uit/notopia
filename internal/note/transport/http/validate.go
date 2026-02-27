package http

import (
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/api"
	"github.com/oapi-codegen/gin-middleware"
)

func ValidateHandler() (gin.HandlerFunc, error) {
	spec, err := api.GetSpec(nil, api.NoteSpec)
	if err != nil {
		return nil, err
	}
	spec.Servers = nil
	return ginmiddleware.OapiRequestValidator(spec), nil
}
