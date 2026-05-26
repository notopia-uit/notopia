package commonconfig

import (
	"net"
	"strconv"
)

type ServerAddress struct {
	Host string `default:"0.0.0.0" json:"host" mapstructure:"host" validate:"omitempty,hostname|ip" yaml:"host"`
	Port uint16 `default:"-"       json:"port" mapstructure:"port" validate:"required,port"         yaml:"port"`
}

func (s *ServerAddress) Address() string {
	return net.JoinHostPort(s.Host, strconv.FormatUint(uint64(s.Port), 10))
}
