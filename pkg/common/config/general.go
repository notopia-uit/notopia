package commonconfig

type AppEnv string

const (
	AppEnvDevelopment AppEnv = "development"
	AppEnvProduction  AppEnv = "production"
)

type General struct {
	AppEnv AppEnv `default:"production" json:"app_env" mapstructure:"app_env" validate:"omitempty,oneof=development production" yaml:"app_env"`
}
