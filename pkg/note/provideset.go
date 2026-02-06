package note

import "github.com/goforj/wire"

var ProviderSet = wire.NewSet(ProvideValidate)
