package app

import (
	"context"

	commonhandler "github.com/notopia-uit/notopia/pkg/common/handler"
)

type GetPolicy struct{}

type GetPolicyHandler struct{ read BillingReadModel }

func NewGetPolicyHandler(read BillingReadModel) *GetPolicyHandler {
	return &GetPolicyHandler{read: read}
}

var ProvideGetPolicyHandler = NewGetPolicyHandler

type GetPolicyQuery commonhandler.Query[GetPolicy, BillingPolicy]

var _ GetPolicyQuery = (*GetPolicyHandler)(nil)

func (h *GetPolicyHandler) Handle(ctx context.Context, _ *GetPolicy) (BillingPolicy, error) {
	p, err := h.read.Policy(ctx)
	if err != nil {
		return BillingPolicy{}, err
	}
	return newBillingPolicy(p), nil
}
