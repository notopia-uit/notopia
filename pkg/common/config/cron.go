package commonconfig

import "time"

type CronJob struct {
	Name     string        `json:"name"     mapstructure:"name"     validate:"required" yaml:"name"`
	Schedule string        `json:"schedule" mapstructure:"schedule" validate:"required" yaml:"schedule"`
	Timeout  time.Duration `json:"timeout"  mapstructure:"timeout"  validate:"required" yaml:"timeout"`
	Enabled  bool          `json:"enabled"  mapstructure:"enabled"  yaml:"enabled"`
}
