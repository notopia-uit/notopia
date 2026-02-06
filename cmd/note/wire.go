//go:build wireinject

package note

import (
	"github.com/goforj/wire"
	"github.com/notopia-uit/notopia/pkg/note/validator"
)

func InitializeApp(cfg *Config) (*App, error) {
	wire.Build(validator.ProviderSet)
	return nil, nil
}
