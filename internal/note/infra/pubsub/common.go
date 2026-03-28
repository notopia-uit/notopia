package pubsub

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

func NewWatermillLogger(logger *slog.Logger) watermill.LoggerAdapter {
	return watermill.NewSlogLogger(logger)
}

var ProvideWatermillLogger = NewWatermillLogger

func NewIntegrationMarshaler() cqrs.CommandEventMarshaler {
	return cqrs.JSONMarshaler{}
}

var ProvideIntegrationMarshaler = NewIntegrationMarshaler

func NewKafkaTracer() kafka.SaramaTracer {
	return kafka.NewOTELSaramaTracer()
}

var ProvideKafkaTracer = NewKafkaTracer

type KafkaPublisher struct {
	kafka.Publisher
}

func NewKafkaPublisher(
	cfg commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	tracer kafka.SaramaTracer,
) (*KafkaPublisher, error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		logger,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}
	return &KafkaPublisher{Publisher: *publisher}, nil
}

var ProvideKafkaPublisher = NewKafkaPublisher
