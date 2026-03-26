package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/notopia-uit/notopia/api"
	"github.com/oapi-codegen/gin-middleware"

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
	App                  *app.App
	ServerURL            string
	WorkspaceEventPubSub app.WorkspaceEventPubSub
}

var _ IStrictHandler = (*StrictHandler)(nil)

func NewStrictHandler(
	app *app.App,
	cfg *config.Server,
	workspaceEventPubSub app.WorkspaceEventPubSub,
) *StrictHandler {
	return &StrictHandler{
		App:                  app,
		ServerURL:            cfg.URL,
		WorkspaceEventPubSub: workspaceEventPubSub,
	}
}

var ProvideStrictHandler = NewStrictHandler

func NewHandler(
	strictServer IStrictHandler,
) IHandler {
	return note.NewStrictHandler(strictServer, []note.StrictMiddlewareFunc{})
}

var ProvideHandler = NewHandler

func ValidateHandler() (gin.HandlerFunc, error) {
	spec, err := api.GetSpec(nil, api.NoteSpec)
	if err != nil {
		return nil, err
	}
	spec.Servers = nil
	return ginmiddleware.OapiRequestValidator(spec), nil
}

func RegisterRoutes(
	e *gin.Engine,
	handler IHandler,
) error {
	validateHandler, err := ValidateHandler()
	if err != nil {
		return err
	}
	api := e.Group("/")
	{
		api.Use(commonhttp.GatewayUserAuth())
		api.Use(validateHandler)
		//exhaustruct:ignore
		options := note.GinServerOptions{
			ErrorHandler: strictServerErrorHandler,
		}
		note.RegisterHandlersWithOptions(api, handler, options)
	}
	e.GET("/ping", func(c *gin.Context) {
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
			logger.ErrorContext(ctx, "failed to shutdown http server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

func (h *HTTP) Run() error {
	return h.ListenAndServe()
}

var Provide = New
