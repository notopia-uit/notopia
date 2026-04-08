package component

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
)

func NewWatermillJsonMarshaler() *cqrs.JSONMarshaler {
	return &cqrs.JSONMarshaler{}
}

var ProvideWatermillJsonMarshaler = NewWatermillJsonMarshaler

func NewWatermillLogger(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermillLogger = NewWatermillLogger
