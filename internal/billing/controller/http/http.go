package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/api"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/config"
	"github.com/notopia-uit/notopia/pkg/api/billing"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"
)

type (
	IHandler       = billing.ServerInterface
	IStrictHandler = billing.StrictServerInterface
)

type StrictHandler struct {
	App           *app.Server
	Identity      *commonconfig.Identity
	webhookSecret string
}

var _ IStrictHandler = (*StrictHandler)(nil)

func NewStrictHandler(app *app.Server, identity *commonconfig.Identity, stripeCfg *config.Stripe) *StrictHandler {
	return &StrictHandler{
		App:           app,
		Identity:      identity,
		webhookSecret: stripeCfg.WebhookSecret.Reveal(),
	}
}

var ProvideStrictHandler = NewStrictHandler

func NewHandler(strictServer IStrictHandler) IHandler {
	options := billing.StrictGinServerOptions{
		RequestErrorHandlerFunc:  strictHandlerRequestErrorHandler,
		HandlerErrorFunc:         strictHandlerError,
		ResponseErrorHandlerFunc: strictHandlerResponseErrorHandler,
	}
	return billing.NewStrictHandlerWithOptions(strictServer, nil, options)
}

var ProvideHandler = NewHandler

func ValidateHandler() (gin.HandlerFunc, error) {
	spec, err := api.GetSpec(nil, api.BillingSpec)
	if err != nil {
		return nil, err
	}
	spec.Servers = nil
	spec.Security = nil
	opts := &ginmiddleware.Options{
		ErrorHandler: ginMiddlewareErrorHandler,
		Options: openapi3filter.Options{
			MultiError:         true,
			AuthenticationFunc: openapi3filter.NoopAuthenticationFunc,
		},
	}
	return ginmiddleware.OapiRequestValidatorWithOptions(spec, opts), nil
}

const rawBodyKey = "billing.rawBody"

func captureRawBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodPost && c.Request.URL.Path == "/billing/webhooks/stripe" {
			body, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.AbortWithStatus(http.StatusBadRequest)
				return
			}
			c.Set(rawBodyKey, body)
			c.Request.Body = io.NopCloser(bytes.NewReader(body))
		}
		c.Next()
	}
}

func rawBody(ctx context.Context) []byte {
	if c, ok := ctx.(*gin.Context); ok {
		if v, ok := c.Get(rawBodyKey); ok {
			if b, ok := v.([]byte); ok {
				return b
			}
		}
	}
	return nil
}

func RegisterRoutes(e *gin.Engine, handler IHandler) error {
	validateHandler, err := ValidateHandler()
	if err != nil {
		return fmt.Errorf("failed to create validate handler: %w", err)
	}
	api := e.Group("/")
	{
		api.Use(commonhttp.GatewayUserAuth())
		api.Use(captureRawBody())
		api.Use(validateHandler)
		billing.RegisterHandlers(api, handler)
	}
	e.GET("/billing/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})
	return nil
}

type HTTP struct {
	*http.Server
}

func New(
	ctx context.Context,
	ginEngine *gin.Engine,
	handler IHandler,
	cfg *config.Server,
	logger *slog.Logger,
) (*HTTP, func(), error) {
	if err := RegisterRoutes(ginEngine, handler); err != nil {
		return nil, nil, err
	}
	server := &HTTP{
		Server: &http.Server{
			Addr:    cfg.HTTP.Address(),
			Handler: ginEngine,
		},
	}
	cleanup := func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown http server", slog.Any("error", err))
		}
	}
	return server, cleanup, nil
}

func (h *HTTP) Run() error {
	if err := h.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

var Provide = New
