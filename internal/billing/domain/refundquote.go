package domain

import (
	"fmt"
	"slices"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type RefundLine struct {
	paymentRef      string
	paidAmount      Money
	amount          Money
	coverageSeconds int64
	unusedSeconds   int64
}

type RefundLineParams struct {
	PaymentRef      string
	PaidAmount      Money
	Amount          Money
	CoverageSeconds int64
	UnusedSeconds   int64
}

func NewRefundLine(p RefundLineParams) (RefundLine, error) {
	if p.PaymentRef == "" {
		return RefundLine{}, errs.NewInvalid("refund line needs a payment reference")
	}
	if p.Amount.currency != p.PaidAmount.currency {
		return RefundLine{}, errs.NewInvalid("refund line currency differs from the paid currency")
	}
	if p.Amount.amount > p.PaidAmount.amount {
		return RefundLine{}, errs.NewInvalid(fmt.Sprintf("refund %d exceeds paid %d on %s", p.Amount.amount, p.PaidAmount.amount, p.PaymentRef))
	}
	return RefundLine{
		paymentRef:      p.PaymentRef,
		paidAmount:      p.PaidAmount,
		amount:          p.Amount,
		coverageSeconds: p.CoverageSeconds,
		unusedSeconds:   p.UnusedSeconds,
	}, nil
}

func (l RefundLine) PaymentRef() string     { return l.paymentRef }
func (l RefundLine) PaidAmount() Money      { return l.paidAmount }
func (l RefundLine) Amount() Money          { return l.amount }
func (l RefundLine) CoverageSeconds() int64 { return l.coverageSeconds }
func (l RefundLine) UnusedSeconds() int64   { return l.unusedSeconds }

// RefundQuote.Amount is the sum of Lines; each line becomes one Refund.
type RefundQuote struct {
	amount        Money
	unusedSeconds int64
	periodSeconds int64
	lines         []RefundLine
}

func NewRefundQuote(lines []RefundLine, unusedSeconds, periodSeconds int64) (RefundQuote, error) {
	var total Money
	for i, l := range lines {
		if i == 0 {
			total = l.amount
			continue
		}
		var err error
		if total, err = total.Add(l.amount); err != nil {
			return RefundQuote{}, err
		}
	}
	return RefundQuote{
		amount:        total,
		unusedSeconds: unusedSeconds,
		periodSeconds: periodSeconds,
		lines:         slices.Clone(lines),
	}, nil
}

func (q RefundQuote) Amount() Money        { return q.amount }
func (q RefundQuote) UnusedSeconds() int64 { return q.unusedSeconds }
func (q RefundQuote) PeriodSeconds() int64 { return q.periodSeconds }
func (q RefundQuote) Lines() []RefundLine  { return slices.Clone(q.lines) }

// FullRefundQuote refunds every payment in full (duplicate checkout).
func FullRefundQuote(payments ...PaymentInfo) (RefundQuote, error) {
	lines := make([]RefundLine, 0, len(payments))
	for _, p := range payments {
		line, err := NewRefundLine(RefundLineParams{
			PaymentRef:      p.paymentRef,
			PaidAmount:      p.amount,
			Amount:          p.amount,
			CoverageSeconds: 0,
			UnusedSeconds:   0,
		})
		if err != nil {
			return RefundQuote{}, err
		}
		lines = append(lines, line)
	}
	return NewRefundQuote(lines, 0, 0)
}
