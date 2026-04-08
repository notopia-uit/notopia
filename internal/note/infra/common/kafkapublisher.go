package common

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
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
) (*KafkaPublisher, error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		logger,
	)
	if err != nil {
		return nil, err
	}
	return &KafkaPublisher{
		Publisher: publisher,
	}, nil
}

var ProvideKafkaPublisher = NewKafkaPublisher

var (
	_ outbox.Publisher               = (*KafkaPublisher)(nil)
	_ integrationpublisher.Publisher = (*KafkaPublisher)(nil)
)
