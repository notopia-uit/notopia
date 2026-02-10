package config

import "fmt"

type ServerAddress struct {
	Host string `json:"host" mapstructure:"host" validate:"omitempty,hostname" yaml:"host"`
	Port uint16 `json:"port" mapstructure:"port" validate:"required,port"      yaml:"port"`
}

func (s *ServerAddress) Address() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}
