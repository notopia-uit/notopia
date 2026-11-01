package pgrepo

import "github.com/notopia-uit/notopia/internal/billing/errs"

func requireTx(inTx bool, op string) error {
	if inTx {
		return nil
	}
	return errs.NewInternal(op+" must run inside a unit of work", nil)
}

func requireLockTx(inTx, forUpdate bool) error {
	if !forUpdate {
		return nil
	}
	return requireTx(inTx, "locking read")
}
