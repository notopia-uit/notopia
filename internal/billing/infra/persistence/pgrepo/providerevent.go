package pgrepo

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type providerEventLog struct{ q *pgsqlc.Queries }

func (l *providerEventLog) MarkProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	n, err := l.q.InsertProviderEvent(ctx, &pgsqlc.InsertProviderEventParams{EventID: eventID, EventType: eventType, ReceivedAt: time.Now()})
	if err != nil {
		return false, wrap("mark provider event processed", err)
	}
	return n == 1, nil
}

func (l *providerEventLog) PruneBefore(ctx context.Context, before time.Time) (int64, error) {
	n, err := l.q.DeleteProviderEventsBefore(ctx, before)
	return n, wrap("prune provider events", err)
}
