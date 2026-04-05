package pg

import (
	"context"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type RepoRegistry struct {
	uow       *UnitOfWork
	txQueries *pgsqlc.Queries

	workspace domain.WorkspaceRepo
	folder    domain.FolderRepo
	note      domain.NoteRepo

	wsOnce     sync.Once
	folderOnce sync.Once
	noteOnce   sync.Once
}

var _ domain.RepoRegistry = (*RepoRegistry)(nil)

func (r *RepoRegistry) Workspace() domain.WorkspaceRepo {
	r.wsOnce.Do(func() {
		r.workspace = NewWorkspace(nil, r.txQueries, true)
	})
	return r.workspace
}

func (r *RepoRegistry) Folder() domain.FolderRepo {
	r.folderOnce.Do(func() {
		r.folder = NewFolder(nil, r.txQueries, true)
	})
	return r.folder
}

func (r *RepoRegistry) Note() domain.NoteRepo {
	r.noteOnce.Do(func() {
		r.note = NewNote(nil, r.txQueries, true)
	})
	return r.note
}

type UnitOfWork struct {
	pool *pgxpool.Pool
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(pool *pgxpool.Pool) *UnitOfWork {
	return &UnitOfWork{
		pool: pool,
	}
}

var ProvideUnitOfWork = NewUnitOfWork

// NOTE: an AI said about chaining error is not a good idea?
func (u *UnitOfWork) Execute(
	ctx context.Context,
	fn func(repoRegistry domain.RepoRegistry) errs.Error,
) (cerr errs.Error) {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to begin transaction", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			cerr = errs.NewPersistenceInternal("failed to rollback transaction", fmt.Errorf("%w: %v", cerr, err))
		}
	}()

	txQueries := pgsqlc.New(tx)
	repoRegistry := &RepoRegistry{
		uow:        u,
		txQueries:  txQueries,
		workspace:  nil,
		folder:     nil,
		note:       nil,
		wsOnce:     sync.Once{},
		folderOnce: sync.Once{},
		noteOnce:   sync.Once{},
	}

	if err := fn(repoRegistry); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction", err)
	}
	return nil
}
