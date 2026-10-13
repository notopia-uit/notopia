package app

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ArchivePrice struct {
	PriceID uuid.UUID
	UserID  string
}

type ArchivePriceHandler struct{ catalog *catalog }

func NewArchivePriceHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *ArchivePriceHandler {
	return &ArchivePriceHandler{catalog: newCatalog(uow, provider, settings)}
}

var ProvideArchivePriceHandler = NewArchivePriceHandler

type ArchivePriceCmd commonhandler.Cmd[ArchivePrice]

var _ ArchivePriceCmd = (*ArchivePriceHandler)(nil)

func (h *ArchivePriceHandler) Handle(ctx context.Context, cmd *ArchivePrice) error {
	var price *domain.Price
	if err := h.catalog.uow.Execute(ctx, func(r domain.RepoRegistry) error {
		var err error
		if price, err = r.Price().GetByID(ctx, cmd.PriceID, true); err != nil {
			return err
		}
		if err := price.Archive(time.Now(), cmd.UserID); err != nil {
			return err
		}
		return r.Price().Save(ctx, price)
	}); err != nil {
		return err
	}
	if err := h.catalog.provider.ArchivePrice(ctx, price.ProviderPriceID()); err != nil {
		return err
	}
	return h.catalog.refreshPortal(ctx, cmd.UserID)
}
