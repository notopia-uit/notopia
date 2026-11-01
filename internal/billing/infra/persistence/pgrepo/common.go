package pgrepo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/billing/domain"
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
