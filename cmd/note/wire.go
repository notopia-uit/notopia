//go:build wireinject

package main

import (
	"context"

	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note"
)

func InitializeServer(ctx context.Context) (*note.Server, error) {
	wire.Build(note.ProviderSet)
	return nil, nil
}
