package config

import "github.com/spf13/viper"

func NewViper() *viper.Viper {
	return viper.NewWithOptions(
		viper.ExperimentalBindStruct(),
	)
}

var ProvideViper = NewViper
