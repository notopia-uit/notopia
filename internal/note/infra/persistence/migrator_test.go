package persistence

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"testing"
	"testing/fstest"

	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestMigratorUpDownStatus(t *testing.T) {
	ctx := context.Background()
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	db.SetMaxOpenConns(1)
	migrations := fstest.MapFS{"00001_create.sql": {Data: []byte("-- +goose Up\nCREATE TABLE example (id INTEGER);\n-- +goose Down\nDROP TABLE example;\n")}}
	provider, err := goose.NewProvider(goose.DialectSQLite3, db, migrations)
	require.NoError(t, err)
	var output bytes.Buffer
	m := NewMigrator(provider, slog.New(slog.NewTextHandler(&output, nil)))
	require.NoError(t, m.Status(ctx))
	require.Contains(t, output.String(), "pending")
	require.NoError(t, m.Up(ctx))
	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT count(*) FROM example").Scan(&count))
	output.Reset()
	require.NoError(t, m.Status(ctx))
	require.Contains(t, output.String(), "applied")
	require.NoError(t, m.Down(ctx))
	require.Error(t, db.QueryRowContext(ctx, "SELECT count(*) FROM example").Scan(&count))
}
