package domain

import (
	"fmt"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

type Money struct {
	amount   int64
	currency Currency
}

func NewMoney(amount int64, currency Currency) (Money, error) {
	if !currency.IsValid() {
		return Money{}, errs.NewInvalid("money needs a valid currency")
	}
	if amount < 0 {
		return Money{}, errs.NewInvalid(fmt.Sprintf("money amount %d must not be negative", amount))
	}
	return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() Currency { return m.currency }

func (m Money) Add(o Money) (Money, error) {
	if m.currency != o.currency {
		return Money{}, errs.NewInvalid(fmt.Sprintf("cannot add %s to %s", o.currency, m.currency))
	}
	return Money{amount: m.amount + o.amount, currency: m.currency}, nil
}

type Payer struct {
	id    string
	email string
	name  string
}

func NewPayer(id, email, name string) (Payer, error) {
	if id == "" {
		return Payer{}, errs.NewInvalid("payer needs an id")
	}
	return Payer{id: id, email: email, name: name}, nil
}

func (p Payer) ID() string    { return p.id }
func (p Payer) Email() string { return p.email }
func (p Payer) Name() string  { return p.name }

func (p Payer) WithContact(email, name string) Payer {
	return Payer{id: p.id, email: email, name: name}
}
