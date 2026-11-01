package pgrepo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

func drain(ctx context.Context, events []domain.Event) {
	for _, e := range events {
		slog.DebugContext(
			ctx,
			"billing domain event",
			slog.String("type", typeName(e)),
			slog.String("aggregate_id", e.GetAggregateID()),
			slog.String("user_id", e.GetUserID()),
		)
	}
}

func typeName(v any) string {
	return fmt.Sprintf("%T", v)
}

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
