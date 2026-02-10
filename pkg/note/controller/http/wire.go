package http

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/common/controller/http"
)

var ProviderSet = wire.NewSet(
	ProvideHandler,
	Provide,
	ProvideStrictHandler,
	ProvideHealthManager,
	http.ProviderSet,
	wire.Bind(new(IStrictHandler), new(*StrictHandler)),
)
