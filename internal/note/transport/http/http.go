package http

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/config"
)

type (
	IHTTPHandler   = note.ServerInterface
	IStrictHandler = note.StrictServerInterface
)

type StrictHandler struct {
	app *app.App
}

func NewStrictHandler(app *app.App) *StrictHandler {
	return &StrictHandler{
		app: app,
	}
}

var _ IStrictHandler = (*StrictHandler)(nil)

var ProvideStrictHandler = NewStrictHandler

func NewHandler(
	strictServer IStrictHandler,
) IHTTPHandler {
	return note.NewStrictHandler(strictServer, []note.StrictMiddlewareFunc{})
}

var ProvideHandler = NewHandler

type Server struct {
	*http.Server
}

func New(
	ctx context.Context,
	ginEngine *gin.Engine,
	httpHandler IHTTPHandler,
	cfg *config.Server,
) (*Server, func(), error) {
	if err := RegisterRoutes(ginEngine, httpHandler); err != nil {
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
			slog.Error("failed to shutdown http server", slog.String("error", err.Error()))
		}
	}
	return server, cleanup, nil
}

func (s *Server) Run() error {
	return s.ListenAndServe()
}

var Provide = New
