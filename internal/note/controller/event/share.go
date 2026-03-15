package event

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func NewWatermillLogger(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermillLogger = NewWatermillLogger

func NewWatermillMarshaler() cqrs.CommandEventMarshaler {
	return cqrs.JSONMarshaler{}
}

var ProvideWatermillMarshaler = NewWatermillMarshaler
