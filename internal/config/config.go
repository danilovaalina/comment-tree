package config

import (
	"os"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Addr         string `mapstructure:"addr"`
	DatabaseURL  string `mapstructure:"db_url"`
	DefaultLimit uint64 `mapstructure:"default_limit"`
	MaxLimit     uint64 `mapstructure:"max_limit"`
}

func Load() (Config, error) {
	v := viper.New()

	v.SetDefault("default_limit", 10)
	v.SetDefault("max_limit", 100)
	v.SetDefault("addr", ":8080")

	v.AddConfigPath(".")
	v.SetConfigName("config")
	v.SetConfigType("yaml")

	err := v.ReadInConfig()
	if err != nil {
		if !errors.As(err, &viper.ConfigFileNotFoundError{}) {
			return Config{}, errors.WithStack(err)
		}
	}

	v.SetConfigFile(".env")
	v.SetConfigType("env")
	err = v.MergeInConfig()
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return Config{}, errors.WithStack(err)
		}
	}

	v.AutomaticEnv()

	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	var cfg Config
	if err = v.Unmarshal(&cfg); err != nil {
		return Config{}, errors.WithDetail(err, "unable to decode into struct")
	}

	return cfg, nil
}
