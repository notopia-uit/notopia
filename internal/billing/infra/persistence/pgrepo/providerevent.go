package pgrepo

import (
	"context"
	"time"

	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/errs"

	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type ProviderEventLog struct {
	q    *pgsqlc.Queries
	inTx bool
}

var _ domain.ProviderEventLog = (*ProviderEventLog)(nil)

func NewProviderEventLog(q *pgsqlc.Queries) *ProviderEventLog {
	return &ProviderEventLog{q: q, inTx: false}
}

var ProvideProviderEventLog = NewProviderEventLog

func newProviderEventLogInTx(q *pgsqlc.Queries) *ProviderEventLog {
	return &ProviderEventLog{q: q, inTx: true}
}

func (l *ProviderEventLog) MarkProcessed(ctx context.Context, eventID, eventType string) (bool, error) {
	if err := requireTx(l.inTx, "mark processed"); err != nil {
		return false, err
	}
	n, err := l.q.InsertProviderEvent(ctx, &pgsqlc.InsertProviderEventParams{EventID: eventID, EventType: eventType, ReceivedAt: time.Now()})
	if err != nil {
		return false, errs.NewInternal("failed to mark provider event processed", err)
	}
	return n == 1, nil
}

func (l *ProviderEventLog) PruneBefore(ctx context.Context, before time.Time) (int64, error) {
	if err := requireTx(l.inTx, "prune"); err != nil {
		return 0, err
	}
	n, err := l.q.DeleteProviderEventsBefore(ctx, before)
	if err != nil {
		return 0, errs.NewInternal("failed to prune provider events", err)
	}
	return n, nil
}
