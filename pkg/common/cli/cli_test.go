package cli

import (
	"context"
	"os"
	"os/exec"
	"testing"

	"github.com/notopia-uit/notopia/pkg/metadata"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

type testRunner struct{ err error }

func (r testRunner) Run(context.Context) error { return r.err }

type testMigrator struct{}

func (testMigrator) Up(context.Context) error     { return nil }
func (testMigrator) Down(context.Context) error   { return nil }
func (testMigrator) Status(context.Context) error { return nil }

type testSeeder struct{}

func (testSeeder) Run(context.Context) error { return nil }

func TestExecuteFailuresExitNonzero(t *testing.T) {
	if os.Getenv("CLI_SUBPROCESS") == "1" {
		root := NewRoot(metadata.ServiceName("test"), "", NewServeCommand(func(context.Context) (Runner, func(), error) {
			return testRunner{err: context.DeadlineExceeded}, func() {}, nil
		}))
		root.SetArgs([]string{"serve"})
		Execute(root)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestExecuteFailuresExitNonzero$")
	cmd.Env = append(os.Environ(), "CLI_SUBPROCESS=1")
	err := cmd.Run()
	require.Error(t, err)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
}

func TestExecuteMissingSubcommandExitsNonzero(t *testing.T) {
	if os.Getenv("CLI_EMPTY_SUBPROCESS") == "1" {
		root := NewRoot(metadata.ServiceName("test"), "")
		root.SetArgs(nil)
		Execute(root)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=^TestExecuteMissingSubcommandExitsNonzero$")
	cmd.Env = append(os.Environ(), "CLI_EMPTY_SUBPROCESS=1")
	err := cmd.Run()
	require.Error(t, err)
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.NotZero(t, exitErr.ExitCode())
}

func TestFailedInitializedCommandCleansUp(t *testing.T) {
	if os.Getenv("CLI_CLEANUP_SUBPROCESS") == "1" {
		root := NewRoot(metadata.ServiceName("test"), "", NewServeCommand(func(context.Context) (Runner, func(), error) {
			return testRunner{err: context.DeadlineExceeded}, func() { os.WriteFile(os.Getenv("CLI_CLEANUP_FILE"), []byte("cleaned"), 0600) }, nil
		}))
		root.SetArgs([]string{"serve"})
		Execute(root)
		return
	}
	file := t.TempDir() + "/cleanup"
	cmd := exec.Command(os.Args[0], "-test.run=^TestFailedInitializedCommandCleansUp$")
	cmd.Env = append(os.Environ(), "CLI_CLEANUP_SUBPROCESS=1", "CLI_CLEANUP_FILE="+file)
	err := cmd.Run()
	var exitErr *exec.ExitError
	require.ErrorAs(t, err, &exitErr)
	require.Equal(t, 1, exitErr.ExitCode())
	content, err := os.ReadFile(file)
	require.NoError(t, err)
	require.Equal(t, "cleaned", string(content))
}

func TestMigrateAndSeedRootSpans(t *testing.T) {
	exporter := tracetest.NewInMemoryExporter()
	provider := trace.NewTracerProvider(trace.WithSyncer(exporter))
	previous := otel.GetTracerProvider()
	otel.SetTracerProvider(provider)
	t.Cleanup(func() { otel.SetTracerProvider(previous); _ = provider.Shutdown(context.Background()) })
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
		want string
	}{
		{"migrate", NewMigrateCommand(func(context.Context) (Migrator, func(), error) { return testMigrator{}, func() {}, nil }), "cli.migrate.up"},
		{"seed", NewSeedCommand(func(context.Context) (Seeder, func(), error) { return testSeeder{}, func() {}, nil }), "cli.seed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := NewRoot(metadata.ServiceName("test"), "", tc.cmd)
			root.SetArgs([]string{tc.name})
			if tc.name == "migrate" {
				root.SetArgs([]string{"migrate", "up"})
			}
			require.NoError(t, root.ExecuteContext(context.Background()))
			spans := exporter.GetSpans()
			require.NotEmpty(t, spans)
			require.Equal(t, tc.want, spans[len(spans)-1].Name)
		})
	}
}
