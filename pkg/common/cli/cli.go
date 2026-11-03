package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/notopia-uit/notopia/pkg/logging"
	"github.com/notopia-uit/notopia/pkg/metadata"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/otel"
)

type Injector[T any] func(context.Context) (T, func(), error)

type Runner interface{ Run(context.Context) error }
type Seeder interface{ Run(context.Context) error }
type Migrator interface {
	Up(context.Context) error
	Down(context.Context) error
	Status(context.Context) error
}

func NewRoot(name metadata.ServiceName, logLevelEnv string, cmds ...*cobra.Command) *cobra.Command {
	root := &cobra.Command{}
	root.Use = string(name)
	root.SilenceUsage = true
	root.SilenceErrors = true
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		if cmd == root {
			_ = cmd.Help()
			return errors.New("a subcommand is required")
		}
		return logging.FirstStart(logLevelEnv)
	}
	root.AddCommand(cmds...)
	return root
}

func NewServeCommand(inject Injector[Runner]) *cobra.Command {
	return &cobra.Command{Use: "serve", RunE: func(cmd *cobra.Command, _ []string) error {
		service, cleanup, err := inject(cmd.Context())
		if err != nil {
			return err
		}
		defer cleanup()
		return service.Run(cmd.Context())
	}}
}

func NewMigrateCommand(inject Injector[Migrator]) *cobra.Command {
	command := &cobra.Command{Use: "migrate"}
	for _, action := range []string{"up", "down", "status"} {
		action := action
		command.AddCommand(&cobra.Command{Use: action, RunE: func(cmd *cobra.Command, _ []string) error {
			migrator, cleanup, err := inject(cmd.Context())
			if err != nil {
				return err
			}
			defer cleanup()
			ctx, span := otel.Tracer("github.com/notopia-uit/notopia/pkg/common/cli").Start(cmd.Context(), "cli.migrate."+action)
			defer span.End()
			switch action {
			case "up":
				return migrator.Up(ctx)
			case "down":
				return migrator.Down(ctx)
			default:
				return migrator.Status(ctx)
			}
		}})
	}
	return command
}

func NewSeedCommand(inject Injector[Seeder]) *cobra.Command {
	return &cobra.Command{Use: "seed", RunE: func(cmd *cobra.Command, _ []string) error {
		seeder, cleanup, err := inject(cmd.Context())
		if err != nil {
			return err
		}
		defer cleanup()
		ctx, span := otel.Tracer("github.com/notopia-uit/notopia/pkg/common/cli").Start(cmd.Context(), "cli.seed")
		defer span.End()
		return seeder.Run(ctx)
	}}
}

func Execute(root *cobra.Command) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	missingCommand := false
	explicitHelp := false
	help := root.HelpFunc()
	root.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		help(cmd, args)
		if cmd == root {
			missingCommand = !cmd.Flags().Changed("help")
		}
	})
	root.SetHelpCommand(&cobra.Command{
		Use: "help [command]",
		RunE: func(cmd *cobra.Command, args []string) error {
			explicitHelp = true
			target, _, err := root.Find(args)
			if err != nil {
				return err
			}
			if target == nil {
				target = root
			}
			return target.Help()
		},
	})
	err := root.ExecuteContext(ctx)
	if err == nil && missingCommand && !explicitHelp {
		err = errors.New("a subcommand is required")
	}
	if err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}
