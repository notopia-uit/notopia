package components

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/controller/domainevent"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
)

var DomainEventPubSubProviderSet = wire.NewSet(
	ProvideDomainPubSub,
	wire.Bind(new(outbox.Publisher), new(*DomainPubSub)),
	wire.Bind(new(domainevent.Subcriber), new(*DomainPubSub)),
)

var ProviderSet = wire.NewSet(
	DomainEventPubSubProviderSet,
	ProvideValidate,
	ProvideWatermillJsonMarshaler,
	ProvideWatermillLogger,
)
