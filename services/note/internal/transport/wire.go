package controller

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/services/note/internal/transport/grpc"
	"github.com/notopia-uit/notopia/services/note/internal/transport/http"
)

var ProviderSet = wire.NewSet(
	http.ProviderSet,
	grpc.ProviderSet,
)
