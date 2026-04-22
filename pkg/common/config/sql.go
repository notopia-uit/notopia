package commonconfig

import (
	"fmt"
	"net/url"

	"github.com/spf13/viper"
)

type SQL struct {
	URL      string `json:"url"      mapstructure:"url"      validate:"omitempty"                                                          yaml:"url"`
	Scheme   string `json:"scheme"   mapstructure:"scheme"   validate:"omitempty,oneof=postgres mysql"                                     yaml:"scheme"`
	Host     string `json:"host"     mapstructure:"host"     validate:"required_without=URL,hostname_rfc1123"                              yaml:"host"`
	Port     uint16 `json:"port"     mapstructure:"port"     validate:"omitempty,min=1,max=65535"                                          yaml:"port"`
	User     string `json:"user"     mapstructure:"user"     validate:""                                                                   yaml:"user"`
	Password string `json:"password" mapstructure:"password" validate:""                                                                   yaml:"password"`
	Name     string `json:"name"     mapstructure:"name"     validate:""                                                                   yaml:"name"`
	SSLMode  string `json:"sslmode"  mapstructure:"sslmode"  validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full" yaml:"sslmode"`
}

func (s *SQL) GetDSN() string {
	if s.URL != "" {
		return s.URL
	}
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

	u := &url.URL{
		Scheme: s.Scheme,
		User:   url.UserPassword(s.User, s.Password),
		Host:   fmt.Sprintf("%s:%d", s.Host, s.Port),
		Path:   s.Name,
	}
	q := u.Query()
	q.Set("sslmode", s.SSLMode)
	u.RawQuery = q.Encode()

	return u.String()
}

func SQLViperSetDefault(
	viper *viper.Viper,
	prefix string,
) {
	viper.SetDefault(prefix+".port", 5432)
}
