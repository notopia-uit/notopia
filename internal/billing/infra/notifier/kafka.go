package notifier

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/billing/app"
	"github.com/notopia-uit/notopia/internal/billing/component"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
)

type Kafka struct {
	publisher message.Publisher
}

var _ app.Notifier = (*Kafka)(nil)

func NewKafka(cfg *commonconfig.Kafka, logger watermill.LoggerAdapter, tracer kafka.SaramaTracer, serviceName metadata.ServiceName) (*Kafka, func(), error) {
	publisher, err := kafka.NewPublisher(kafka.PublisherConfig{Brokers: cfg.Brokers, Tracer: tracer}, logger)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create kafka publisher: %w", err)
	}
	p := wotel.NewNamedPublisherDecorator(serviceName.String(), publisher)
	return &Kafka{publisher: p}, func() { _ = p.Close() }, nil
}

var ProvideKafka = NewKafka

func (k *Kafka) publish(ctx context.Context, topic string, payload any) {
	body, err := json.Marshal(payload)
	if err == nil {
		err = k.publisher.Publish(topic, message.NewMessageWithContext(ctx, watermill.NewUUID(), body))
	}
	if err != nil {
		slog.WarnContext(
			ctx,
			"billing notification not published",
			slog.String("topic", topic),
			slog.Any("error", err),
		)
	}
}

func (k *Kafka) SubscriptionChanged(ctx context.Context, workspaceID uuid.UUID, plan domain.PlanTier, status domain.Status) {
	sharePlan, err := toSharePlan(plan)
	if err == nil {
		var shareStatus string
		if shareStatus, err = toShareStatus(status); err == nil {
			k.publish(ctx, component.TopicSubscriptionChanged, &share.SubscriptionChangedEvent{
				WorkspaceId: workspaceID,
				Plan:        sharePlan,
				Status:      shareStatus,
			})
			return
		}
	}
	slog.WarnContext(
		ctx,
		"billing notification not published",
		slog.String("topic", component.TopicSubscriptionChanged),
		slog.Any("error", err),
	)
}
