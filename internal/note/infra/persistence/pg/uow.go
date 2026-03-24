package pg

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/notopia-uit/notopia/internal/note/domain"
	"github.com/notopia-uit/notopia/internal/note/errs"
	"github.com/notopia-uit/notopia/internal/note/infra/persistence/pgsqlc"
)

type RepoRegistry struct {
	workspace *Workspace
	folder    *Folder
	note      *Note
}

var _ domain.RepoRegistry = (*RepoRegistry)(nil)

func (r *RepoRegistry) Workspace() domain.WorkspaceRepo { return r.workspace }

func (r *RepoRegistry) Folder() domain.FolderRepo { return r.folder }

func (r *RepoRegistry) Note() domain.NoteRepo { return r.note }

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

func (u *UnitOfWork) Execute(ctx context.Context, fn func(repoRegistry domain.RepoRegistry) errs.Error) (cerr errs.Error) {
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
	repoRegistry := &RepoRegistry{
		workspace: NewWorkspace(txQueries, tx),
		folder:    NewFolder(txQueries, tx),
		note:      NewNote(txQueries, tx),
	}

	if err := fn(repoRegistry); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return errs.NewPersistenceInternal("failed to commit transaction", err)
	}
	return nil
}
