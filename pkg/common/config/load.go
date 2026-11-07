package commonconfig

import (
	"errors"
	"fmt"
	"log/slog"
	"reflect"
	"strings"

	"github.com/creasty/defaults"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

func Load[T any](validate *validator.Validate, v *viper.Viper, envPrefix, fileName string) (*T, error) {
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.SetConfigName(fileName)
	v.AddConfigPath(".")
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err == nil {
		slog.Info("configuration loaded", slog.String("file", v.ConfigFileUsed()))
	} else {
		var notFound viper.ConfigFileNotFoundError
		if !errors.As(err, &notFound) {
			return nil, fmt.Errorf("cannot read config file: %w", err)
		}
	}
	cfg := new(T)
	if err := defaults.Set(cfg); err != nil {
		return nil, fmt.Errorf("cannot set default config: %w", err)
	}
	if err := bindEnvFields(v, reflect.TypeOf(*cfg), ""); err != nil {
		return nil, fmt.Errorf("cannot bind config environment: %w", err)
	}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config from env or config file: %w", err)
	}
	slog.Info("configuration", slog.Any("config", cfg))
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}
	return cfg, nil
}

func bindEnvFields(v *viper.Viper, typ reflect.Type, prefix string) error {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return nil
	}
	for i := range typ.NumField() {
		field := typ.Field(i)
		if !field.IsExported() {
			continue
		}
		parts := strings.Split(field.Tag.Get("mapstructure"), ",")
		name := parts[0]
		if name == "-" {
			continue
		}
		squash := len(parts) > 1 && parts[1] == "squash"
		if name == "" && !squash {
			name = field.Name
		}
		key := prefix
		if name != "" {
			if key != "" {
				key += "."
			}
			key += name
		}
		ft := field.Type
		if ft.Kind() == reflect.Pointer {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct {
			if err := bindEnvFields(v, ft, key); err != nil {
				return err
			}
		} else if err := v.BindEnv(key); err != nil {
			return err
		}
	}
	return nil
}
