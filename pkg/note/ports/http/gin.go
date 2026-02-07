package http

import "github.com/gin-gonic/gin"

func NewGin() *gin.Engine {
	return gin.Default()
}

var ProvideGin = NewGin
