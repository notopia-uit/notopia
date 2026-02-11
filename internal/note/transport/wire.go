package controller

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/transport/grpc"
	"github.com/notopia-uit/notopia/internal/note/transport/http"
)

var ProviderSet = wire.NewSet(
	http.ProviderSet,
	grpc.ProviderSet,
)
