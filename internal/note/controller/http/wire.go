package http

import (
	"github.com/goforj/wire"
	commonhttp "github.com/notopia-uit/notopia/pkg/common/http"
)

var ProviderSet = wire.NewSet(
	Provide,
	ProvideHandler,
	ProvideHealthManager,
	wire.Struct(new(StrictHandler), "*"),
	commonhttp.ProviderSet,
	wire.Bind(new(IStrictHandler), new(*StrictHandler)),
)
