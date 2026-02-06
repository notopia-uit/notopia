package config

import (
	"github.com/goforj/wire"
	"github.com/spf13/viper"
)

var ProvideSet = wire.NewSet(
	ProvideConfig,
	viper.New,
)
