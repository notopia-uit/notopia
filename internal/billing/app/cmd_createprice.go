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

func NewCreatePriceHandler(uow domain.UnitOfWork, prices domain.PriceRepo, policies domain.BillingPolicyRepo, provider PaymentProvider, settings *Settings) *CreatePriceHandler {
	return &CreatePriceHandler{catalog: newCatalog(uow, prices, policies, provider, settings)}
}

var ProvideCreatePriceHandler = NewCreatePriceHandler

type CreatePriceCmd commonhandler.Cmd[CreatePrice]

var _ CreatePriceCmd = (*CreatePriceHandler)(nil)

func (h *CreatePriceHandler) Handle(ctx context.Context, cmd *CreatePrice) error {
	currency, err := domain.ParseCurrency(cmd.Currency)
	if err != nil {
		return err
	}
	amount, err := domain.NewMoney(cmd.Amount, currency)
	if err != nil {
		return err
	}
	return h.catalog.createPrice(
		ctx,
		cmd.ID,
		toDomainPlanTier(cmd.Plan),
		toDomainInterval(cmd.Interval),
		amount,
		cmd.UserID,
	)
}
