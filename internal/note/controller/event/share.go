package event

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
)

func NewWatermillLogger(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermillLogger = NewWatermillLogger
