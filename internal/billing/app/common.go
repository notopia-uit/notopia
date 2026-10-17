package app

import (
	"context"
	"errors"
	"log/slog"

	"github.com/notopia-uit/notopia/internal/billing/errs"
)

func hasCode(err error, code errs.Code) bool {
	var e errs.Error
	return errors.As(err, &e) && e.Code() == code
}

// cancelAtProvider runs after the local cancel committed. A failure is logged, not returned:
// the local cancel wins and reconcile retries the provider cancel.
func cancelAtProvider(ctx context.Context, provider PaymentProvider, providerSubscriptionID string) {
	if err := provider.CancelSubscriptionNow(ctx, providerSubscriptionID); err != nil {
		slog.ErrorContext(
			ctx, "cancelling subscription at the provider failed; reconcile will retry",
			slog.String("provider_subscription_id", providerSubscriptionID),
			slog.Any("error", err),
		)
	}
}
