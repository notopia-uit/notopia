package note

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/notopia-uit/notopia/pkg/note/config"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

type Server struct {
	ginEngine  *gin.Engine
	connectRpc pbconnect.NoteServiceHandler
	Config     *config.Server
	server     *http.Server
}

func ProvideServer(
	ginEngine *gin.Engine,
	connectRpc pbconnect.NoteServiceHandler,
	cfg *config.Config,
) *Server {
	s := &Server{
		ginEngine:  ginEngine,
		connectRpc: connectRpc,
		Config:     cfg.Server,
	}
	path, handler := pbconnect.NewNoteServiceHandler(connectRpc)
	s.ginEngine.Any(path+"*", gin.WrapH(handler))
	s.server = &http.Server{
		Addr:    cfg.Server.Address(),
		Handler: s.ginEngine,
	}
	return s
}

func (s *Server) Run() error {
	if err := s.server.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func ProvideGinEngine() *gin.Engine {
	router := gin.Default()
	return router
}
