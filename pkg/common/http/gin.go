package commonhttp

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
	sloggin "github.com/samber/slog-gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type GinSlogHandlerFunc gin.HandlerFunc

func NewGinSlogHandler(
	logCfg *commonconfig.Log,
	logger *slog.Logger,
) GinSlogHandlerFunc {
	cfg := sloggin.Config{
		WithUserAgent:      true,
		WithRequestID:      true,
		WithRequestBody:    false,
		WithRequestHeader:  true,
		WithResponseBody:   false,
		WithResponseHeader: true,
		WithSpanID:         true,
		WithTraceID:        true,
	}
	return GinSlogHandlerFunc(sloggin.NewWithConfig(logger, cfg))
}

var ProvideGinSlogHandler = NewGinSlogHandler

type OtelGinHandlerFunc gin.HandlerFunc

func NewOtelGinHandler(
	serviceName metadata.ServiceName,
) OtelGinHandlerFunc {
	return OtelGinHandlerFunc(otelgin.Middleware(
		serviceName.String(),
	))
}

var ProvideOtelGinHandler = NewOtelGinHandler

func NewGin(
	slogHandler GinSlogHandlerFunc,
	otelHandler OtelGinHandlerFunc,
	generalCfg *commonconfig.General,
) *gin.Engine {
	if generalCfg.AppEnv == commonconfig.AppEnvProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.HandlerFunc(otelHandler))
	r.Use(gin.HandlerFunc(slogHandler))
	return r
}

var ProvideGin = NewGin
