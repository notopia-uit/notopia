package app

import (
	"context"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type ListPrices struct{}

type ListPricesHandler struct{ read BillingReadModel }

func NewListPricesHandler(read BillingReadModel) *ListPricesHandler {
	return &ListPricesHandler{read: read}
}

var ProvideListPricesHandler = NewListPricesHandler

type ListPricesQuery commonhandler.Query[ListPrices, []Price]

var _ ListPricesQuery = (*ListPricesHandler)(nil)

func (h *ListPricesHandler) Handle(ctx context.Context, _ *ListPrices) ([]Price, error) {
	prices, err := h.read.Prices(ctx)
	if err != nil {
		return nil, err
	}
	return newPrices(prices), nil
}
