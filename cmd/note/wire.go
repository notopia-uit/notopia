//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note"
)

func InitializeServer(ctx context.Context) (*note.Server, func(), error) {
	panic(wire.Build(note.ProviderSet))
}
