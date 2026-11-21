package commonconfig

import (
	"fmt"
	"net/url"
	"strings"
)

type SQL struct {
	URL      Secret `default:"-"        json:"url"      mapstructure:"url"      validate:"omitempty,url,startswith=postgres"                                  yaml:"url"`
	Scheme   string `default:"postgres" json:"scheme"   mapstructure:"scheme"   validate:"omitempty,oneof=postgres postgresql"                                yaml:"scheme"`
	Host     string `default:"-"        json:"host"     mapstructure:"host"     validate:"required_without=URL,omitempty,hostname_rfc1123"                    yaml:"host"`
	Port     uint16 `default:"5432"     json:"port"     mapstructure:"port"     validate:"omitempty,min=1,max=65535"                                          yaml:"port"`
	User     string `default:"-"        json:"user"     mapstructure:"user"     validate:""                                                                   yaml:"user"`
	Password Secret `default:"-"        json:"password" mapstructure:"password" validate:""                                                                   yaml:"password"`
	Name     string `default:"-"        json:"name"     mapstructure:"name"     validate:""                                                                   yaml:"name"`
	SSLMode  string `default:"-"        json:"sslmode"  mapstructure:"sslmode"  validate:"omitempty,oneof=disable allow prefer require verify-ca verify-full" yaml:"sslmode"`
}

func (s *SQL) GetDSN() string {
	if s.URL != "" {
		return s.URL.Reveal()
	}
	escape := strings.NewReplacer(`\`, `\\`, `'`, `\'`)
	return fmt.Sprintf(
		"host='%s' port=%d user='%s' password='%s' dbname='%s' sslmode='%s'",
		escape.Replace(s.Host),
		s.Port,
		escape.Replace(s.User),
		escape.Replace(s.Password.Reveal()),
		escape.Replace(s.Name),
		escape.Replace(s.SSLMode),
	)
}

func (s *SQL) GetURL() string {
	if s.URL != "" {
		return s.URL.Reveal()
	}

	u := &url.URL{
		Scheme: s.Scheme,
		User:   url.UserPassword(s.User, s.Password.Reveal()),
		Host:   fmt.Sprintf("%s:%d", s.Host, s.Port),
		Path:   s.Name,
	}
	if s.SSLMode != "" {
		u.RawQuery = url.Values{"sslmode": {s.SSLMode}}.Encode()
	}

	return u.String()
}
