package domain

import (
	"github.com/goforj/wire"
)

var ProviderSet = wire.NewSet(
	ProvideNoteService,
	ProvideTrashService,
)
