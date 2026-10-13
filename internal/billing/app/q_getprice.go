package app

import (
	"context"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/errs"
	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetPrice struct {
	ID uuid.UUID
}

type GetPriceHandler struct{ read BillingReadModel }

func NewGetPriceHandler(read BillingReadModel) *GetPriceHandler {
	return &GetPriceHandler{read: read}
}

var ProvideGetPriceHandler = NewGetPriceHandler

type GetPriceQuery commonhandler.Query[GetPrice, Price]

var _ GetPriceQuery = (*GetPriceHandler)(nil)

// Handle scans every price: the admin catalog holds a handful of rows.
func (h *GetPriceHandler) Handle(ctx context.Context, q *GetPrice) (Price, error) {
	prices, err := h.read.Prices(ctx)
	if err != nil {
		return Price{}, err
	}
	for _, p := range prices {
		if p.ID() == q.ID {
			return newPrice(p), nil
		}
	}
	return Price{}, errs.NewNotFound("price not found")
}
