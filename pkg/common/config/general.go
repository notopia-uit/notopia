package config

type General struct {
	TZ string `json:"tz" mapstructure:"tz" validate:"" yaml:"tz"`
}
