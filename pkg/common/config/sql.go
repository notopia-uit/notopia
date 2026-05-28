package commonconfig

import (
	"fmt"
	"net/url"
)

type SQL struct {
	URL      string `default:"-"    json:"url"      mapstructure:"url"      validate:"omitempty"                                                          yaml:"url"`
	Scheme   string `default:"-"    json:"scheme"   mapstructure:"scheme"   validate:"omitempty,oneof=postgres mysql"                                     yaml:"scheme"`
	Host     string `default:"5432" json:"host"     mapstructure:"host"     validate:"required_without=URL,omitempty,hostname_rfc1123"                    yaml:"host"`
	Port     uint16 `default:"-"    json:"port"     mapstructure:"port"     validate:"omitempty,min=1,max=65535"                                          yaml:"port"`
	User     string `default:"-"    json:"user"     mapstructure:"user"     validate:""                                                                   yaml:"user"`
	Password string `default:"-"    json:"password" mapstructure:"password" validate:""                                                                   yaml:"password"`
	Name     string `default:"-"    json:"name"     mapstructure:"name"     validate:""                                                                   yaml:"name"`
	SSLMode  string `default:"-"    json:"sslmode"  mapstructure:"sslmode"  validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full" yaml:"sslmode"`
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
