package domain

import (
	"time"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

// PaymentInfo is one paid invoice payment. CoverageStart/End is the service window it paid
// for (invoice line periods); zero for lifetime.
type PaymentInfo struct {
	amount        Money
	paymentRef    string
	paidAt        time.Time
	coverageStart time.Time
	coverageEnd   time.Time
}

type PaymentInfoParams struct {
	Amount        Money
	PaymentRef    string
	PaidAt        time.Time
	CoverageStart time.Time
	CoverageEnd   time.Time
}

func NewPaymentInfo(p PaymentInfoParams) (PaymentInfo, error) {
	if p.PaymentRef == "" {
		return PaymentInfo{}, errs.NewInvalid("payment needs a reference")
	}
	if !p.Amount.currency.IsValid() {
		return PaymentInfo{}, errs.NewInvalid("payment needs an amount with a currency")
	}
	if p.CoverageEnd.Before(p.CoverageStart) {
		return PaymentInfo{}, errs.NewInvalid("payment coverage ends before it starts")
	}
	return PaymentInfo{
		amount:        p.Amount,
		paymentRef:    p.PaymentRef,
		paidAt:        p.PaidAt,
		coverageStart: p.CoverageStart,
		coverageEnd:   p.CoverageEnd,
	}, nil
}

func (p PaymentInfo) Amount() Money            { return p.amount }
func (p PaymentInfo) PaymentRef() string       { return p.paymentRef }
func (p PaymentInfo) PaidAt() time.Time        { return p.paidAt }
func (p PaymentInfo) CoverageStart() time.Time { return p.coverageStart }
func (p PaymentInfo) CoverageEnd() time.Time   { return p.coverageEnd }
