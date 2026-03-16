package pubsub

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func NewWatermillLogger(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermillLogger = NewWatermillLogger

func NewIntegrationMarshaler() cqrs.CommandEventMarshaler {
	return cqrs.JSONMarshaler{}
}

var ProvideIntegrationMarshaler = NewIntegrationMarshaler
