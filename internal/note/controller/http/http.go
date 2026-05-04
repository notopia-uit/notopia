package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/notopia-uit/notopia/api"
	ginmiddleware "github.com/oapi-codegen/gin-middleware"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

type (
	IHandler       = note.ServerInterface
	IStrictHandler = note.StrictServerInterface
)

type StrictHandler struct {
	App               *app.Server
	BaseURL           *url.URL
	WorkspaceEventHub app.WorkspaceEventHub
}

var _ IStrictHandler = (*StrictHandler)(nil)

func NewStrictHandler(
	app *app.Server,
	cfg *config.Server,
	workspaceEventHub app.WorkspaceEventHub,
) (*StrictHandler, error) {
	baseURL, err := url.Parse(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base url: %w", err)
	}
	return &StrictHandler{
		App:               app,
		BaseURL:           baseURL,
		WorkspaceEventHub: workspaceEventHub,
	}, nil
}

var ProvideStrictHandler = NewStrictHandler

func NewHandler(
	strictServer IStrictHandler,
) IHandler {
	return note.NewStrictHandler(strictServer, nil)
}

var ProvideHandler = NewHandler

func ValidateHandler() (gin.HandlerFunc, error) {
	spec, err := api.GetSpec(nil, api.NoteSpec)
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

func RegisterRoutes(
	e *gin.Engine,
	handler IHandler,
) error {
	validateHandler, err := ValidateHandler()
	if err != nil {
		return fmt.Errorf("failed to create validate handler: %w", err)
	}
	api := e.Group("/")
	{
		api.Use(commonhttp.GatewayUserAuth())
		api.Use(validateHandler)
		api.Use(StrictHandlerErrorMiddleware())
		//exhaustruct:ignore
		options := note.GinServerOptions{
			ErrorHandler: serverErrorHandler,
		}
		note.RegisterHandlersWithOptions(api, handler, options)
	}
	e.GET("/note/ping", func(c *gin.Context) {
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
	return h.ListenAndServe()
}

var Provide = New
