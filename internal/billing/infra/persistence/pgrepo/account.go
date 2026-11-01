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

type accountRepo struct{ q *pgsqlc.Queries }

func (r *accountRepo) GetByWorkspace(ctx context.Context, workspaceID uuid.UUID, forUpdate bool) (*domain.BillingAccount, error) {
	row, err := r.q.GetBillingAccount(ctx, pgsqlc.GetBillingAccountParams{WorkspaceID: workspaceID, ForUpdate: forUpdate})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errs.NewInternal("failed to get billing account", err)
	}
	return pgmapper.ToAccount(row)
}

func (r *accountRepo) ListByPayer(ctx context.Context, payerID string) ([]*domain.BillingAccount, error) {
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

func (r *accountRepo) Save(ctx context.Context, a *domain.BillingAccount) error {
	p := a.Payer()
	err := r.q.UpsertBillingAccount(ctx, &pgsqlc.UpsertBillingAccountParams{
		WorkspaceID: a.WorkspaceID(), ProviderCustomerID: a.ProviderCustomerID(),
		PayerID: p.ID(), PayerEmail: p.Email(), PayerName: p.Name(), PayerMissingSince: a.PayerMissingSince(),
		CreatedAt: a.CreatedAt(), UpdatedAt: a.UpdatedAt(),
	})
	drain(ctx, a.PopEvents())
	if err != nil {
		return errs.NewInternal("failed to save billing account", err)
	}
	return nil
}
