package commonconfig

import (
	"net"
	"strconv"

	"github.com/spf13/viper"
)

type ServerAddress struct {
	Host string `json:"host" mapstructure:"host" validate:"omitempty,hostname|ip" yaml:"host"`
	Port uint16 `json:"port" mapstructure:"port" validate:"required,port"         yaml:"port"`
}

func (s *ServerAddress) Address() string {
	return net.JoinHostPort(s.Host, strconv.FormatUint(uint64(s.Port), 10))
}

func ServerAddressViperSetDefault(
	viper *viper.Viper,
	prefix string,
	port uint16,
) {
	viper.SetDefault(prefix+".host", "0.0.0.0")
	viper.SetDefault(prefix+".port", port)
}
