package app

import (
	"context"
	"log/slog"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

// Stripe retries a webhook for up to 3 days; 30 days keeps a wide margin for manual resends.
const providerEventRetention = 30 * 24 * time.Hour

type PruneProviderEvents struct{}

type PruneProviderEventsHandler struct {
	uow domain.UnitOfWork
}

func NewPruneProviderEventsHandler(uow domain.UnitOfWork) *PruneProviderEventsHandler {
	return &PruneProviderEventsHandler{uow: uow}
}

var ProvidePruneProviderEventsHandler = NewPruneProviderEventsHandler

type PruneProviderEventsCmd commonhandler.Cmd[PruneProviderEvents]

var _ PruneProviderEventsCmd = (*PruneProviderEventsHandler)(nil)

func (h *PruneProviderEventsHandler) Handle(ctx context.Context, _ *PruneProviderEvents) error {
	return h.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		pruned, err := r.ProviderEventLog().PruneBefore(ctx, time.Now().Add(-providerEventRetention))
		if err != nil {
			return err
		}
		slog.InfoContext(ctx, "pruned processed provider events", slog.Int64("count", pruned))
		return nil
	})
}
