package commonconfig

import (
	"net"
	"strconv"
)

type ServerAddress struct {
	Host string `json:"host" mapstructure:"host" validate:"omitempty,hostname|ip" yaml:"host"`
	Port int    `json:"port" mapstructure:"port" validate:"required,port"         yaml:"port"`
}

func (s *ServerAddress) Address() string {
	return net.JoinHostPort(s.Host, strconv.Itoa(s.Port))
}
