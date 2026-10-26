package pgrepo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"
)

func notFound(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

func wrap(msg string, err error) error {
	if err == nil {
		return nil
	}
	return errs.NewInternal(msg, err)
}

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
