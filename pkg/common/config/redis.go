package commonconfig

import "github.com/redis/go-redis/v9"

type Redis struct {
	Addr     string `default:"-" json:"addr"     mapstructure:"addr"     validate:"required" yaml:"addr"`
	Username string `default:"-" json:"username" mapstructure:"username" validate:""         yaml:"username"`
	Password Secret `default:"-" json:"password" mapstructure:"password" validate:""         yaml:"password"`
	DB       int    `default:"0" json:"db"       mapstructure:"db"       validate:"min=0"    yaml:"db"`
}

func (r *Redis) Options() *redis.Options {
	return &redis.Options{Addr: r.Addr, Username: r.Username, Password: r.Password.Reveal(), DB: r.DB}
}
