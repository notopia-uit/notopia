package app

import (
	"context"

	"github.com/google/uuid"
)

type IntegrationPublisher interface {
	Publish(ctx context.Context, event ...IntegrationEvent) error
}

type IntegrationEvent interface {
	isIntegrationEvent()
}

type IntegrationEventUserWorkspaceRoleUpdated struct {
	WorkspaceID uuid.UUID
	UserID      string
	Role        WorkspaceRole
}

var _ IntegrationEvent = (*IntegrationEventUserWorkspaceRoleUpdated)(nil)

func (e IntegrationEventUserWorkspaceRoleUpdated) isIntegrationEvent() {}

type IntegrationEventWorkspaceMemberAdded struct {
	WorkspaceID uuid.UUID
	UserID      string
	Role        WorkspaceRole
}

var _ IntegrationEvent = (*IntegrationEventWorkspaceMemberAdded)(nil)

func (e IntegrationEventWorkspaceMemberAdded) isIntegrationEvent() {}

type IntegrationEventWorkspaceMemberRemoved struct {
	WorkspaceID uuid.UUID
	UserID      string
}

var _ IntegrationEvent = (*IntegrationEventWorkspaceMemberRemoved)(nil)

func (e IntegrationEventWorkspaceMemberRemoved) isIntegrationEvent() {}
