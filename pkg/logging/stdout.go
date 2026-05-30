package logging

import (
	"log/slog"
	"os"

	commonconfig "github.com/notopia-uit/notopia/pkg/common/config"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type StdoutHandler slog.Handler

func NewStdoutHandler(
	logCfg *commonconfig.Log,
	generalCfg *commonconfig.General,
) StdoutHandler {
	if !logCfg.Enabled {
		return slog.DiscardHandler
	}
	if logCfg.Pretty {
		zl := log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
		return zerolog.NewSlogHandler(zl)
	}
	return slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: logCfg.GetSlogLevel(),
	})
}

var ProvideStdoutHandler = NewStdoutHandler
