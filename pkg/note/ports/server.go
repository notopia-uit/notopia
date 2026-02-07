package ports

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/pkg/note/config"
	"github.com/notopia-uit/notopia/pkg/note/ports/rpc"
)

type Server struct {
	server *http.Server
}

func NewServer(
	ginEngine *gin.Engine,
	httpHandler note.ServerInterface,
	httpRpcHandler *rpc.HTTPServiceHandler,
	logger *slog.Logger,
	cfg *config.Config,
) *Server {
	slog.SetDefault(logger)

	s := &Server{}
	note.RegisterHandlers(ginEngine, httpHandler)
	ginEngine.Any(httpRpcHandler.Path+"*", gin.WrapH(httpRpcHandler.Handler))

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
