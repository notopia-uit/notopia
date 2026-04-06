package components

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/notopia-uit/notopia/internal/note/controller/domainevent"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
)

type DomainPubSub struct {
	*gochannel.GoChannel
}

var (
	_ domainevent.Subcriber = (*DomainPubSub)(nil)
	_ outbox.Publisher      = (*DomainPubSub)(nil)
)

func NewDomainPubSub(
	logger watermill.LoggerAdapter,
) *DomainPubSub {
	return &DomainPubSub{
		GoChannel: gochannel.NewGoChannel(gochannel.Config{}, logger),
	}
}

var ProvideDomainPubSub = NewDomainPubSub

func NewWatermillJsonMarshaler() *cqrs.JSONMarshaler {
	return &cqrs.JSONMarshaler{}
}

var ProvideWatermillJsonMarshaler = NewWatermillJsonMarshaler

func NewWatermillLogger(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermillLogger = NewWatermillLogger
