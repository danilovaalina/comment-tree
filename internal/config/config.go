package config

import (
	"os"

	"github.com/cockroachdb/errors"
	"github.com/spf13/viper"
)

type Config struct {
	Addr         string `mapstructure:"ADDR"`
	DatabaseURL  string `mapstructure:"DB_URL"`
	DefaultLimit uint64 `mapstructure:"DEFAULT_LIMIT"`
	MaxLimit     uint64 `mapstructure:"MAX_LIMIT"`
}

func Load() (Config, error) {
	v := viper.New()

	v.SetDefault("DEFAULT_LIMIT", 10)
	v.SetDefault("MAX_LIMIT", 100)
	v.SetDefault("ADDR", ":8080")

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

	//_ = v.BindEnv("addr", "ADDR")
	//_ = v.BindEnv("db_url", "DB_URL")

	var cfg Config
	if err = v.Unmarshal(&cfg); err != nil {
		return Config{}, errors.WithDetail(err, "unable to decode into struct")
	}

	return cfg, nil
}
