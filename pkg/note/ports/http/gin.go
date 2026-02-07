package http

import (
	"log/slog"

	ginslog "github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"
)

type GinSlogHandlerFunc = gin.HandlerFunc

func NewSlogHandler(
	logger *slog.Logger,
) GinSlogHandlerFunc {
	return ginslog.SetLogger(
		ginslog.WithLogger(
			func(c *gin.Context, _ *slog.Logger) *slog.Logger {
				return logger.With("user_id", c.GetString("x-forwarded-id"))
			},
		),
		ginslog.WithSkipPath([]string{
			"/healthz",
			"/readyz",
			"/metrics",
			"/livez",
		}),
	)
}

var ProvideGinSlog = NewSlogHandler

func NewGin(
	slogHandler GinSlogHandlerFunc,
) *gin.Engine {
	r := gin.Default()
	r.Use(slogHandler)
	return r
}

var ProvideGin = NewGin
