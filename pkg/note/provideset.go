package note

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note/components"
	"github.com/notopia-uit/notopia/pkg/note/config"
	"github.com/notopia-uit/notopia/pkg/pb/pbconnect"
)

var ServerSet = wire.NewSet(
	wire.Struct(new(pbconnect.UnimplementedNoteServiceHandler), "*"),
	wire.Bind(
		new(pbconnect.NoteServiceHandler),
		new(*pbconnect.UnimplementedNoteServiceHandler),
	),
	ProvideServer,
	ProvideGinEngine,
)

var ProviderSet = wire.NewSet(
	ServerSet,
	components.ProvideSet,
	config.ProvideSet,
)
