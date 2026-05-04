package common

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/controller/event"
	"github.com/notopia-uit/notopia/internal/note/infra/integrationpublisher"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
)

var KafkaPublisherProviderSet = wire.NewSet(
	ProvideKafkaPublisher,
	wire.Bind(new(event.Publisher), new(*KafkaPublisher)),
	wire.Bind(new(integrationpublisher.Publisher), new(*KafkaPublisher)),
	wire.Bind(new(outbox.Publisher), new(*KafkaPublisher)),
)

var ProviderSet = wire.NewSet(
	KafkaPublisherProviderSet,
)
