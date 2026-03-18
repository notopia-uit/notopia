package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/notopia-uit/notopia/api"
	"github.com/oapi-codegen/gin-middleware"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/notopia-uit/notopia/internal/note/config"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonerror "github.com/notopia-uit/notopia/pkg/common/error"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

type (
	IHandler       = note.ServerInterface
	IStrictHandler = note.StrictServerInterface
)

type StrictHandler struct {
	App            *app.App
	WorkspaceEvent pubsub.WorkspaceEvent
}

var _ IStrictHandler = (*StrictHandler)(nil)

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

func strictServerErrorHandler(c *gin.Context, err error, statusCode int) {
	message := err.Error()
	code := ""
	if domainErr, ok := errors.AsType[*commonerror.Err](err); ok {
		message, code, statusCode = commonerror.ToHTTP(domainErr)
	}

	response := note.Error{
		Code:    code,
		Message: message,
	}

	c.JSON(statusCode, response)
}

type Server struct {
	*http.Server
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
		note.RegisterHandlersWithOptions(api, handler, note.GinServerOptions{
			ErrorHandler: strictServerErrorHandler,
		})
	}
	return nil
}

func New(
	ctx context.Context,
	ginEngine *gin.Engine,
	handler IHandler,
	cfg *config.Server,
	logger *slog.Logger,
) (*Server, func(), error) {
	if err := RegisterRoutes(ginEngine, handler); err != nil {
		return nil, nil, err
	}

	server := &Server{
		Server: &http.Server{
			Addr:    cfg.HTTP.Address(),
			Handler: ginEngine,
		},
	}
	cleanup := func() {
		if err := server.Shutdown(ctx); err != nil {
			logger.ErrorContext(ctx, "failed to shutdown http server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

func (s *Server) Run() error {
	return s.ListenAndServe()
}

var Provide = New
