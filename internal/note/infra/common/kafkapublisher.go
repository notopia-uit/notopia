package common

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/note/controller/event"
	"github.com/notopia-uit/notopia/internal/note/infra/integrationpublisher"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
)

type KafkaPublisher struct {
	message.Publisher
}

func NewKafkaPublisher(
	cfg *commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	tracer kafka.SaramaTracer,
	serviceName metadata.ServiceName,
) (*KafkaPublisher, func(), error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		logger,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}
	otelPublisher := wotel.NewNamedPublisherDecorator(serviceName.String(), publisher)
	cleanup := func() {
		if err := otelPublisher.Close(); err != nil {
			slog.Error("failed to close Kafka publisher", slog.Any("error", err))
		}
	}
	return &KafkaPublisher{
		Publisher: otelPublisher,
	}, cleanup, nil
}

var ProvideKafkaPublisher = NewKafkaPublisher

var (
	_ outbox.Publisher               = (*KafkaPublisher)(nil)
	_ integrationpublisher.Publisher = (*KafkaPublisher)(nil)
	_ event.Publisher                = (*KafkaPublisher)(nil)
)
