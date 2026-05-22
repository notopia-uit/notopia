package pgreadmodel

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/notopia-uit/notopia/internal/note/app"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type GetFolder struct {
	queries *pgsqlc.Queries
}

var _ app.GetFolderReadModel = (*GetFolder)(nil)

func NewGetFolder(queries *pgsqlc.Queries) *GetFolder {
	return &GetFolder{queries: queries}
}

var ProvideGetFolder = NewGetFolder

func (h *GetFolder) Handle(ctx context.Context, p *app.GetFolderReadModelParams) (app.Folder, error) {
	folder, err := h.queries.ReadGetFolder(ctx, pgsqlc.ReadGetFolderParams{
		ID:             p.ID,
		OnlyNonTrashed: p.ExcludeTrashed, // What so ... inconsistency
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return app.Folder{}, errs.NewFolderNotFound(p.ID, err)
		}
		return app.Folder{}, toErr(err)
	}
	return toAppFolder(folder)
}
