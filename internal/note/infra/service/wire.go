package service

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var ProviderSet = wire.NewSet(
	ProvideAuthorization,
	wire.Bind(new(app.Authorization), new(*Authorization)),
)
