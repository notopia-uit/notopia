package http

import (
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/pkg/note/app"
)

type strictServer struct {
	app *app.App
}

func NewStrictServer(app *app.App) *strictServer {
	return &strictServer{
		app: app,
	}
}

var _ note.StrictServerInterface = (*strictServer)(nil)

var ProvideStrictServer = NewStrictServer

func NewServer(
	strictServer note.StrictServerInterface,
	// middlewares []note.StrictMiddlewareFunc, // TODO: we will pass raw here, not that type, and convert to type later?
) note.ServerInterface {
	return note.NewStrictHandler(strictServer, []note.StrictMiddlewareFunc{})
}

var ProvideServer = NewServer
