package infra

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v3/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/notopia-uit/notopia/internal/authorization/app"
	"github.com/notopia-uit/notopia/pkg/api/share"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type IntegrationPublisher struct {
	publisher message.Publisher
}

func NewIntegrationPublisher(
	cfg *commonconfig.Kafka,
	logger watermill.LoggerAdapter,
	tracer kafka.SaramaTracer,
) (*IntegrationPublisher, error) {
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
	return &IntegrationPublisher{
		publisher: publisher,
	}, nil
}

var ProvideIntegrationPublisher = NewIntegrationPublisher

var _ app.IntegrationPublisher = (*IntegrationPublisher)(nil)

func (p *IntegrationPublisher) Publish(ctx context.Context, events ...app.IntegrationEvent) error {
	for _, event := range events {
		transformedEvent, ok := p.transformIntegrationEvent(event)
		if !ok {
			return fmt.Errorf("cannot convert event to integration event: %T", event)
		}
		topic, ok := p.getIntegrationEventTopic(event)
		if !ok {
			return fmt.Errorf("cannot get topic for integration event: %T", event)
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

func (p *IntegrationPublisher) transformIntegrationEvent(event app.IntegrationEvent) (any, bool) {
	switch e := event.(type) {
	case app.IntegrationEventUserWorkspaceRoleUpdated:
		return &share.UserWorkspaceRoleUpdatedEvent{
			WorkspaceId: &e.WorkspaceID,
			UserId:      e.UserID,
			Role:        p.toWorkspaceRole(e.Role),
		}, true
	case app.IntegrationEventWorkspaceMemberAdded:
		return &share.WorkspaceMemberAddedEvent{
			WorkspaceId: &e.WorkspaceID,
			UserId:      e.UserID,
			Role:        p.toWorkspaceRole(e.Role),
		}, true
	case app.IntegrationEventWorkspaceMemberRemoved:
		return &share.WorkspaceMemberRemovedEvent{
			WorkspaceId: &e.WorkspaceID,
			UserId:      e.UserID,
		}, true
	}
	return nil, false
}

func (p *IntegrationPublisher) toWorkspaceRole(role app.WorkspaceRole) share.WorkspaceRole {
	switch role {
	case app.WorkspaceRoleOwner:
		return share.Owner
	case app.WorkspaceRoleEditor:
		return share.Editor
	case app.WorkspaceRoleViewer:
		return share.Viewer
	default:
		return share.Viewer
	}
}

func (p *IntegrationPublisher) getIntegrationEventTopic(event app.IntegrationEvent) (string, bool) {
	switch event.(type) {
	case app.IntegrationEventUserWorkspaceRoleUpdated:
		return "events.integration.authorization.user_workspace_role_updated", true
	case app.IntegrationEventWorkspaceMemberAdded:
		return "events.integration.authorization.workspace_member_added", true
	case app.IntegrationEventWorkspaceMemberRemoved:
		return "events.integration.authorization.workspace_member_removed", true
	}
	return "", false
}
