package pg

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type RepoRegistry struct {
	uow       *UnitOfWork
	txQueries *pgsqlc.Queries
	tx        *sql.Tx

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
		r.workspace = NewWorkspace(nil, r.txQueries, r.tx, true)
	})
	return r.workspace
}

func (r *RepoRegistry) Folder() domain.FolderRepo {
	r.folderOnce.Do(func() {
		r.folder = NewFolder(nil, r.txQueries, r.tx, true)
	})
	return r.folder
}

func (r *RepoRegistry) Note() domain.NoteRepo {
	r.noteOnce.Do(func() {
		r.note = NewNote(nil, r.txQueries, r.tx, true)
	})
	return r.note
}

type UnitOfWork struct {
	queries *pgsqlc.Queries
	sdb     *sql.DB
}

var _ domain.UnitOfWork = (*UnitOfWork)(nil)

func NewUnitOfWork(queries *pgsqlc.Queries, sdb *sql.DB) *UnitOfWork {
	return &UnitOfWork{
		queries: queries,
		sdb:     sdb,
	}
}

var ProvideUnitOfWork = NewUnitOfWork

// NOTE: an AI said about chaining error is not a good idea?
func (u *UnitOfWork) Execute(
	ctx context.Context,
	fn func(repoRegistry domain.RepoRegistry) errs.Error,
) (cerr errs.Error) {
	conn, err := u.sdb.Conn(ctx)
	if err != nil {
		return errs.NewPersistenceInternal("failed to get connection from pool", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			cerr = errs.NewPersistenceInternal("failed to close connection", fmt.Errorf("%w: %v", cerr, err))
		}
	}()

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return errs.NewPersistenceInternal("failed to begin transaction", err)
	}
	defer func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			cerr = errs.NewPersistenceInternal("failed to rollback transaction", fmt.Errorf("%w: %v", cerr, err))
		}
	}()

	var pgxConn *pgx.Conn
	err = conn.Raw(func(driverConn any) error {
		pgxConn = driverConn.(*stdlib.Conn).Conn()
		return nil
	})
	if err != nil {
		return errs.NewPersistenceInternal("failed to get raw connection", err)
	}

	txQueries := pgsqlc.New(pgxConn)
	// NOTE: passing nil to pgxpool because pgxpool in repo used for starting a transaction
	// but in this case transaction is already started in unit of work
	repoRegistry := &RepoRegistry{
		uow:        u,
		txQueries:  txQueries,
		tx:         tx,
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

	if err := tx.Commit(); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction", err)
	}
	return nil
}
