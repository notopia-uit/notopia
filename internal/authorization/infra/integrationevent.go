package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	wotel "github.com/nkonev/watermill-opentelemetry/pkg/opentelemetry"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/notopia-uit/notopia/pkg/metadata"
)

type IntegrationPublisher struct {
	publisher message.Publisher
}

func NewIntegrationPublisher(
	cfg *commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	tracer kafka.SaramaTracer,
	serviceName metadata.ServiceName,
) (*IntegrationPublisher, func(), error) {
	publisher, err := kafka.NewPublisher(
		kafka.PublisherConfig{
			Brokers: cfg.Brokers,
			Tracer:  tracer,
		},
		logger,
	)
	otelPublisher := wotel.NewNamedPublisherDecorator(serviceName.String(), publisher)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Kafka publisher: %w", err)
	}
	cleanup := func() {
		if err := otelPublisher.Close(); err != nil {
			slog.Error("failed to close Kafka publisher", slog.Any("error", err))
		}
	}
	return &IntegrationPublisher{
		publisher: otelPublisher,
	}, cleanup, nil
}

var ProvideIntegrationPublisher = NewIntegrationPublisher

var _ app.IntegrationPublisher = (*IntegrationPublisher)(nil)

func (p *IntegrationPublisher) Publish(ctx context.Context, events ...app.IntegrationEvent) error {
	for _, event := range events {
		transformedEvent, err := p.transformIntegrationEvent(event)
		if err != nil {
			return fmt.Errorf("cannot convert event to integration event: %w", err)
		}
		topic, err := p.getIntegrationEventTopic(event)
		if err != nil {
			return fmt.Errorf("cannot get topic for integration event: %w", err)
		}
		payload, err := json.Marshal(transformedEvent)
		if err != nil {
			return fmt.Errorf("failed to marshal integration event: %w", err)
		}
		msg := message.NewMessageWithContext(ctx, watermill.NewUUID(), payload)
		if err := p.publisher.Publish(topic, msg); err != nil {
			return fmt.Errorf("failed to publish integration event: %w", err)
		}
	}
	return nil
}

func (p *IntegrationPublisher) transformIntegrationEvent(event app.IntegrationEvent) (any, error) {
	switch e := event.(type) {
	case app.IntegrationEventUserWorkspaceRoleUpdated:
		role, err := p.toWorkspaceRole(e.Role)
		if err != nil {
			return nil, fmt.Errorf("failed to convert workspace role within integration event: %w", err)
		}
		return &share.UserWorkspaceRoleUpdatedEvent{
			WorkspaceId: &e.WorkspaceID,
			UserId:      e.UserID,
			Role:        role,
		}, nil
	case app.IntegrationEventWorkspaceMemberAdded:
		role, err := p.toWorkspaceRole(e.Role)
		if err != nil {
			return nil, fmt.Errorf("failed to convert workspace role within integration event: %w", err)
		}
		return &share.WorkspaceMemberAddedEvent{
			WorkspaceId: &e.WorkspaceID,
			UserId:      e.UserID,
			Role:        role,
		}, nil
	case app.IntegrationEventWorkspaceMemberRemoved:
		return &share.WorkspaceMemberRemovedEvent{
			WorkspaceId: &e.WorkspaceID,
			UserId:      e.UserID,
		}, nil
	}
	return nil, fmt.Errorf("unknown integration event type: %T", event)
}

func (p *IntegrationPublisher) toWorkspaceRole(role app.WorkspaceRole) (share.WorkspaceRole, error) {
	switch role {
	case app.WorkspaceRoleOwner:
		return share.Owner, nil
	case app.WorkspaceRoleEditor:
		return share.Editor, nil
	case app.WorkspaceRoleViewer:
		return share.Viewer, nil
	default:
		return share.WorkspaceRole(""), fmt.Errorf("unknown workspace role: %v", role)
	}
}

func (p *IntegrationPublisher) getIntegrationEventTopic(event app.IntegrationEvent) (string, error) {
	switch event.(type) {
	case app.IntegrationEventUserWorkspaceRoleUpdated:
		return "events.integration.authorization.user_workspace_role_updated", nil
	case app.IntegrationEventWorkspaceMemberAdded:
		return "events.integration.authorization.workspace_member_added", nil
	case app.IntegrationEventWorkspaceMemberRemoved:
		return "events.integration.authorization.workspace_member_removed", nil
	}
	return "", fmt.Errorf("unknown integration event type: %T", event)
}
