package app

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/app/command"
	"github.com/notopia-uit/notopia/internal/note/app/event"
	"github.com/notopia-uit/notopia/internal/note/app/pubsub"
	"github.com/notopia-uit/notopia/internal/note/app/query"
)

var ProviderSet = wire.NewSet(
	command.ProviderSet,
	query.ProviderSet,
	event.ProviderSet,
	pubsub.ProviderSet,
	ProvideApp,
)
