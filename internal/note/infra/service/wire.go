package service

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app/service"
)

var ProviderSet = wire.NewSet(
	ProvideAuthorization,
	wire.Bind(new(service.Authorization), new(*Authorization)),
)
