package http

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/pkg/note/config"
)

type (
	IHTTPHandler       = note.ServerInterface
	IStrictHTTPHandler = note.StrictServerInterface
)

type Server struct {
	server *http.Server
}

func NewServer(
	ginEngine *gin.Engine,
	httpHandler IHTTPHandler,
	rpcHTTPHandlerRegister *RPCHTTPHandlerRegister,
	logger *slog.Logger,
	cfg *config.Config,
) *Server {
	slog.SetDefault(logger)

	s := &Server{}
	note.RegisterHandlers(ginEngine, httpHandler)
	ginEngine.Any(
		rpcHTTPHandlerRegister.Path+"*",
		gin.WrapH(rpcHTTPHandlerRegister.Handler),
	)

	s.server = &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: ginEngine,
	}
	return s
}

func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

var ProvideServer = NewServer
