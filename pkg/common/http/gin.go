package commonhttp

import (
	"log/slog"

	ginslog "github.com/gin-contrib/slog"
	"github.com/gin-gonic/gin"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

type GinSlogHandlerFunc gin.HandlerFunc

func NewGinSlogHandler(
	logCfg *commonconfig.Log,
	logger *slog.Logger,
) GinSlogHandlerFunc {
	return GinSlogHandlerFunc(ginslog.SetLogger(
		ginslog.WithLogger(
			func(c *gin.Context, _ *slog.Logger) *slog.Logger {
				logger := logger
				userId := c.GetHeader("X-Forwarded-ID")
				if userId != "" {
					logger = logger.With("user_id", userId)
				}
				return logger
			},
		), ginslog.WithSkipper(func(c *gin.Context) bool {
			switch logCfg.GetSlogLevel() {
			case slog.LevelDebug, slog.LevelInfo:
				return false
			case slog.LevelWarn:
				return c.Writer.Status() < 400
			case slog.LevelError:
				return c.Writer.Status() < 500
			default:
				return true
			}
		}),
	))
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
) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.HandlerFunc(otelHandler))
	r.Use(gin.HandlerFunc(slogHandler))
	return r
}

var ProvideGin = NewGin
