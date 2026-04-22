package logging

import (
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
)

func NewWatermill(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermill = NewWatermill
