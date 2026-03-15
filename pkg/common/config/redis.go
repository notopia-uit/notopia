package commonconfig

type Redis struct {
	Addr string `json:"addr" mapstructure:"addr" validate:"required" yaml:"addr"`
}
