package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type CreatePrice struct {
	ID       uuid.UUID
	Plan     PlanTier
	Interval BillingInterval
	Amount   int64
	Currency string
	UserID   string
}

type CreatePriceHandler struct{ catalog *catalog }

func NewCreatePriceHandler(uow domain.UnitOfWork, provider PaymentProvider, settings *Settings) *CreatePriceHandler {
	return &CreatePriceHandler{catalog: newCatalog(uow, provider, settings)}
}

var ProvideCreatePriceHandler = NewCreatePriceHandler

type CreatePriceCmd commonhandler.Cmd[CreatePrice]

var _ CreatePriceCmd = (*CreatePriceHandler)(nil)

func (h *CreatePriceHandler) Handle(ctx context.Context, cmd *CreatePrice) error {
	return h.catalog.createPrice(
		ctx,
		cmd.ID,
		toDomainPlanTier(cmd.Plan),
		toDomainInterval(cmd.Interval),
		domain.Money{
			Amount:   cmd.Amount,
			Currency: cmd.Currency,
		},
		cmd.UserID,
	)
}
