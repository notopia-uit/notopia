package components

import (
	"github.com/goforj/wire"
)

var ProvideSet = wire.NewSet(
	ProvideValidate,
)
