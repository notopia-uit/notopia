package server

import (
	"github.com/notopia-uit/notopia/pkg/api/note"
	"github.com/notopia-uit/notopia/pkg/note/app"
)

type StrictHTTPHandler struct {
	app *app.App
}

func NewStrictHTTPHandler(app *app.App) *StrictHTTPHandler {
	return &StrictHTTPHandler{
		app: app,
	}
}

var _ IStrictHTTPHandler = (*StrictHTTPHandler)(nil)

var ProvideStrictHTTPHandler = NewStrictHTTPHandler

func NewHTTPHandler(
	strictServer IStrictHTTPHandler,
	// middlewares []note.StrictMiddlewareFunc, // TODO: we will pass raw here, not that type, and convert to type later?
) IHTTPHandler {
	return note.NewStrictHandler(strictServer, []note.StrictMiddlewareFunc{})
}

var ProvideHTTPHandler = NewHTTPHandler
