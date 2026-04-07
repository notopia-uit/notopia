package common

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/internal/note/infra/integrationpublisher"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
)

var KafkaPublisherProviderSet = wire.NewSet(
	ProvideKafkaPublisher,
	wire.Bind(new(outbox.Publisher), new(*KafkaPublisher)),
	wire.Bind(new(integrationpublisher.Publisher), new(*KafkaPublisher)),
)

var ProviderSet = wire.NewSet(
	KafkaPublisherProviderSet,
)
