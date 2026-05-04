package commonconfig

type Meilisearch struct {
	// This include scheme before
	Host   string `json:"host"    mapstructure:"host"    validate:"required,url" yaml:"host"`
	APIKey string `json:"api_key" mapstructure:"api_key" validate:"required"     yaml:"api_key"`
}
