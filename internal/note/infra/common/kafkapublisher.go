package common

import (
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/notopia-uit/notopia/internal/note/controller/event"
	"github.com/notopia-uit/notopia/internal/note/infra/integrationpublisher"
	"github.com/notopia-uit/notopia/internal/note/infra/outbox"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type KafkaPublisher struct {
	*kafka.Publisher
}

func NewKafkaPublisher(
	cfg *commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	tracer kafka.SaramaTracer,
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
	cleanup := func() {
		if err := publisher.Close(); err != nil {
			slog.Error("failed to close Kafka publisher", slog.Any("error", err))
		}
	}
	return &KafkaPublisher{
		Publisher: publisher,
	}, cleanup, nil
}

var ProvideKafkaPublisher = NewKafkaPublisher

var (
	_ outbox.Publisher               = (*KafkaPublisher)(nil)
	_ integrationpublisher.Publisher = (*KafkaPublisher)(nil)
	_ event.Publisher                = (*KafkaPublisher)(nil)
)
