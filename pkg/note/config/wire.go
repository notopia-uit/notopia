package config

import (
	"github.com/goforj/wire"
	"github.com/spf13/viper"
)

var ProviderSet = wire.NewSet(
	ProvideConfig,
	viper.New,
)
