package app

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/pkg/api/note"
	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
)

type BillingChanged struct {
	WorkspaceID uuid.UUID
	Plan        PlanTier
}

type BillingChangedHandler struct {
	publisher WorkspaceEventPublisher
	identity  *commonconfig.Identity
}

func NewBillingChangedHandler(
	publisher WorkspaceEventPublisher,
	identity *commonconfig.Identity,
) *BillingChangedHandler {
	return &BillingChangedHandler{publisher: publisher, identity: identity}
}

var ProvideBillingChangedHandler = NewBillingChangedHandler

func (h *BillingChangedHandler) Handle(ctx context.Context, ev *BillingChanged) error {
	err := h.publisher.Publish(ctx, ev.WorkspaceID, h.identity.ServiceAccountUserID, &WorkspaceEventBillingChanged{
		workspaceEvent[note.WorkspaceBillingChangedEventEvent]{
			Id:    uuid.Must(uuid.NewV7()),
			Event: note.WorkspaceBillingChangedEventEventWorkspaceBillingChangedEvent,
			Data: note.WorkspaceBillingChangedEventData{
				WorkspaceId: &ev.WorkspaceID,
				Plan:        ev.Plan.String(),
			},
		},
	})
	if err != nil {
		return err
	}
	slog.InfoContext(ctx, "Billing changed event published", slog.String("workspace_id", ev.WorkspaceID.String()))
	return nil
}
