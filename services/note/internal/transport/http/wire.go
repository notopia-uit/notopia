package http

import (
	"github.com/goforj/wire"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

var ProviderSet = wire.NewSet(
	ProvideHandler,
	Provide,
	ProvideStrictHandler,
	ProvideHealthManager,
	commonhttp.ProviderSet,
	wire.Bind(new(IStrictHandler), new(*StrictHandler)),
)
