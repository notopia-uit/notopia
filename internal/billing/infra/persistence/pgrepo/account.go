package pgrepo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/billing/errs"

	"github.com/google/uuid"
	"github.com/notopia-uit/notopia/internal/billing/domain"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgmapper"
	"github.com/notopia-uit/notopia/internal/billing/infra/persistence/pgsqlc"
)

type AccountRepo struct {
	q         *pgsqlc.Queries
	publisher EventPublisher
	inTx      bool
}

var _ domain.BillingAccountRepo = (*AccountRepo)(nil)

func NewAccountRepo(q *pgsqlc.Queries) *AccountRepo {
	return &AccountRepo{q: q, publisher: nil, inTx: false}
}

var ProvideAccountRepo = NewAccountRepo

func newAccountRepoInTx(q *pgsqlc.Queries, publisher EventPublisher) *AccountRepo {
	return &AccountRepo{q: q, publisher: publisher, inTx: true}
}

func (r *AccountRepo) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*domain.BillingAccount, error) {
	if err := requireLockTx(r.inTx, forUpdate); err != nil {
		return nil, err
	}
	row, err := r.q.GetBillingAccount(ctx, pgsqlc.GetBillingAccountParams{WorkspaceID: workspaceID, ForUpdate: forUpdate})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get billing account", err)
	}
	return pgmapper.ToAccount(row)
}

func (r *AccountRepo) ListByPayer(ctx context.Context, payerID string) ([]*domain.BillingAccount, error) {
	rows, err := r.q.ListBillingAccountsByPayer(ctx, payerID)
	if err != nil {
		return nil, errs.NewInternal("failed to list billing accounts by payer", err)
	}
	out := make([]*domain.BillingAccount, 0, len(rows))
	for _, row := range rows {
		a, err := pgmapper.ToAccount(row)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, nil
}

func (r *AccountRepo) Save(ctx context.Context, a *domain.BillingAccount) error {
	if err := requireTx(r.inTx, "save"); err != nil {
		return err
	}
	p := a.Payer()
	err := r.q.UpsertBillingAccount(ctx, &pgsqlc.UpsertBillingAccountParams{
		WorkspaceID: a.WorkspaceID(), ProviderCustomerID: a.ProviderCustomerID(),
		PayerID: p.ID(), PayerEmail: p.Email(), PayerName: p.Name(), PayerMissingSince: a.PayerMissingSince(),
		CreatedAt: a.CreatedAt(), UpdatedAt: a.UpdatedAt(),
	})
	if err != nil {
		return errs.NewInternal("failed to save billing account", err)
	}
	return r.publisher.Publish(ctx, a.PopEvents()...)
}
