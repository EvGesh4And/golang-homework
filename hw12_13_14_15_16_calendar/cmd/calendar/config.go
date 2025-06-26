package main

import (
	"github.com/BurntSushi/toml"
	"github.com/caarlos0/env/v10"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf  `toml:"logger" env-prefix:"LOGGER_"`
	Storage StorageConf `toml:"storage" env-prefix:"STORAGE_"`
	HTTP    HTTPConf    `toml:"http" env-prefix:"HTTP_"`
	GRPC    GRPCConf    `toml:"grpc" env-prefix:"GRPC_"`
}

type LoggerConf struct {
	Mod   string `toml:"mod" env:"MOD"`
	Path  string `toml:"path" env:"PATH"`
	Level string `toml:"level" env:"LEVEL"`
}

type StorageConf struct {
	Mod       string `toml:"mod" env:"MOD"`
	DSN       string `toml:"dsn" env:"DSN"`
	Migration string `toml:"migration" env:"MIGRATION"`
}

type HTTPConf struct {
	Host string `toml:"host" env:"HOST"`
	Port int    `toml:"port" env:"PORT"`
}

type GRPCConf struct {
	Host string `toml:"host" env:"HOST"`
	Port int    `toml:"port" env:"PORT"`
}

func NewConfig() (Config, error) {
	var cfg Config
	if _, err := toml.DecodeFile(configFile, &cfg); err != nil {
		return Config{}, err
	}
	if err := env.Parse(&cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}
