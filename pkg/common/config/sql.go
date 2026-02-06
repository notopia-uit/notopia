package config

import "fmt"

type SQL struct {
	DSN      string `json:"dsn"      mapstructure:"dsn"      validate:"omitempty,startswith=postgres://"                                   yaml:"dsn"`
	Host     string `json:"host"     mapstructure:"host"     validate:"required_without=DSN,omitempty,hostname_rfc1123"                    yaml:"host"`
	Port     int    `json:"port"     mapstructure:"port"     validate:"omitempty,min=1,max=65535"                                          yaml:"port"`
	User     string `json:"user"     mapstructure:"user"     validate:""                                                                   yaml:"user"`
	Password string `json:"password" mapstructure:"password" validate:""                                                                   yaml:"password"`
	DBName   string `json:"dbname"   mapstructure:"dbname"   validate:""                                                                   yaml:"dbname"`
	SSLMode  string `json:"sslmode"  mapstructure:"sslmode"  validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full" yaml:"sslmode"`
}

func (pg *SQL) GetDSN() string {
	if pg.DSN != "" {
		return pg.DSN
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		pg.Host,
		pg.Port,
		pg.User,
		pg.Password,
		pg.DBName,
		pg.SSLMode,
	)
}
