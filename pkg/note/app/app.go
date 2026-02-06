package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type App struct {
	GinServer  *gin.Engine
	GRPCServer *http.Server
}
