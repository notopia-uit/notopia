package controller

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/controller/event"
	"github.com/notopia-uit/notopia/internal/note/controller/grpc"
	"github.com/notopia-uit/notopia/internal/note/controller/http"
)

var ProviderSet = wire.NewSet(
	event.ProviderSet,
	grpc.ProviderSet,
	http.ProviderSet,
)
