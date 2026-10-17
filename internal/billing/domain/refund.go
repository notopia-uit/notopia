package domain

import (
	"time"

	"github.com/google/uuid"
)

type Refund struct {
	id               uuid.UUID
	providerRefundID *string
	paymentRef       string
	amount           Money
	reason           RefundReason
	status           RefundStatus
	failureMessage   *string
	createdAt        time.Time
	updatedAt        time.Time
}

type UnmarshalRefundParams struct {
	ID               uuid.UUID
	ProviderRefundID *string
	PaymentRef       string
	Amount           Money
	Reason           RefundReason
	Status           RefundStatus
	FailureMessage   *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func UnmarshalRefund(p UnmarshalRefundParams) *Refund {
	return &Refund{
		id: p.ID, providerRefundID: p.ProviderRefundID, paymentRef: p.PaymentRef, amount: p.Amount,
		reason: p.Reason, status: p.Status, failureMessage: p.FailureMessage, createdAt: p.CreatedAt, updatedAt: p.UpdatedAt,
	}
}

func (r *Refund) ID() uuid.UUID             { return r.id }
func (r *Refund) ProviderRefundID() *string { return r.providerRefundID }
func (r *Refund) PaymentRef() string        { return r.paymentRef }
func (r *Refund) Amount() Money             { return r.amount }
func (r *Refund) Reason() RefundReason      { return r.reason }
func (r *Refund) Status() RefundStatus      { return r.status }
func (r *Refund) FailureMessage() *string   { return r.failureMessage }
func (r *Refund) CreatedAt() time.Time      { return r.createdAt }
func (r *Refund) UpdatedAt() time.Time      { return r.updatedAt }
