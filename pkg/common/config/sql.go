package config

import (
	"fmt"

	"github.com/spf13/viper"
)

type SQL struct {
	URL      string `json:"url"      mapstructure:"url"      validate:"omitempty"                                                          yaml:"url"`
	Host     string `json:"host"     mapstructure:"host"     validate:"required_without=URL,omitempty,hostname_rfc1123"                    yaml:"host"`
	Port     uint16 `json:"port"     mapstructure:"port"     validate:"omitempty,min=1,max=65535"                                          yaml:"port"`
	User     string `json:"user"     mapstructure:"user"     validate:""                                                                   yaml:"user"`
	Password string `json:"password" mapstructure:"password" validate:""                                                                   yaml:"password"`
	Name     string `json:"name"     mapstructure:"name"     validate:""                                                                   yaml:"name"`
	SSLMode  string `json:"sslmode"  mapstructure:"sslmode"  validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full" yaml:"sslmode"`
}

func (s *SQL) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		s.Host,
		s.Port,
		s.User,
		s.Password,
		s.Name,
		s.SSLMode,
	)
}

func (s *SQL) GetURL() string {
	if s.URL != "" {
		return s.URL
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		s.User,
		s.Password,
		s.Host,
		s.Port,
		s.Name,
		s.SSLMode,
	)
}

func SQLViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".port", 5432)
}
