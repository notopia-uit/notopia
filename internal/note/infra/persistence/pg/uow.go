package pg

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type RepoRegistry struct {
	workspace *Workspace
	folder    *Folder
	note      *Note
}

var _ domain.RepoRegistry = (*RepoRegistry)(nil)

func (r *RepoRegistry) Workspace() domain.WorkspaceRepo {
	return r.workspace
}

func (r *RepoRegistry) Folder() domain.FolderRepo {
	return r.folder
}

func (r *RepoRegistry) Note() domain.NoteRepo {
	return r.note
}

type UnitOfWork struct {
	queries *pgsqlc.Queries
	pool    *pgxpool.Pool
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func (u *UnitOfWork) Execute(ctx context.Context, fn func(repoRegistry domain.RepoRegistry) error) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			slog.WarnContext(ctx, "failed to rollback transaction", slog.String("error", err.Error()))
		}
	}()

	repoRegistry := &RepoRegistry{
		workspace: &Workspace{queries: u.queries},
		folder:    &Folder{queries: u.queries},
		note:      &Note{queries: u.queries},
	}

	if err := fn(repoRegistry); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
