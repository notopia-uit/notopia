package controller

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note/controller/grpc"
	"github.com/notopia-uit/notopia/pkg/note/controller/http"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	http.ProviderSet,
	grpc.ProviderSet,
)
