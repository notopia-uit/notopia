package identity

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app"
)

var AuthentikProviderSet = wire.NewSet(
	ProvideAuthentik,
	wire.Bind(new(app.IdentitySvc), new(*Authentik)),
)
