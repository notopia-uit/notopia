package ports

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note/ports/http"
	"github.com/notopia-uit/notopia/pkg/note/ports/rpc"
)

var ProviderSet = wire.NewSet(
	ProvideServer,
	http.ProviderSet,
	rpc.ProviderSet,
)
