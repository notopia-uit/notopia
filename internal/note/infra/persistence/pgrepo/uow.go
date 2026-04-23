package pgrepo

import (
	"context"
	"log/slog"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type RepoRegistry struct {
	uow       *UnitOfWork
	txQueries *pgsqlc.Queries
	publisher Publisher
	runInTx   *RunInTx

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
		r.workspace = NewWorkspace(nil, r.txQueries, r.publisher, r.runInTx, true)
	})
	return r.workspace
}

func (r *RepoRegistry) Folder() domain.FolderRepo {
	r.folderOnce.Do(func() {
		r.folder = NewFolder(nil, r.txQueries, r.publisher, r.runInTx, true)
	})
	return r.folder
}

func (r *RepoRegistry) Note() domain.NoteRepo {
	r.noteOnce.Do(func() {
		r.note = NewNote(nil, r.txQueries, r.publisher, r.runInTx, true)
	})
	return r.note
}

type UnitOfWork struct {
	pool             *pgxpool.Pool
	publisherFactory PublisherFactory
	runInTx          *RunInTx
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(
	pool *pgxpool.Pool,
	publisherFactory PublisherFactory,
	runInTx *RunInTx,
) *UnitOfWork {
	return &UnitOfWork{
		pool:             pool,
		publisherFactory: publisherFactory,
		runInTx:          runInTx,
	}
}

var ProvideUnitOfWork = NewUnitOfWork

// NOTE: an AI said about chaining error is not a good idea?
func (u *UnitOfWork) Execute(
	ctx context.Context,
	fn func(repoRegistry domain.RepoRegistry) error,
) error {
	tx, err := u.pool.Begin(ctx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to begin transaction in RunInTx", err)
	}
	defer func() {
		if p := recover(); p != nil {
			if err = tx.Rollback(ctx); err != nil {
				slog.ErrorContext(
					ctx, "failed to rollback transaction after panic in unit of work",
					slog.Any("error", err),
				)
			}
			panic(p)
		} else if err != nil {
			if err = tx.Rollback(ctx); err != nil {
				slog.ErrorContext(
					ctx, "failed to rollback transaction after error in unit of work",
					slog.Any("error", err),
				)
			}
		}
	}()

	txQueries := pgsqlc.New(tx)
	publisher, err := u.publisherFactory.Create(ctx, tx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to create publisher in unit of work", err)
	}
	repoRegistry := &RepoRegistry{
		uow:        u,
		runInTx:    u.runInTx,
		txQueries:  txQueries,
		publisher:  publisher,
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
		return errs.NewPersistenceInternal("failed to commit transaction in unit of work", err)
	}
	return nil
}
