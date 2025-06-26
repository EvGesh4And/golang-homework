package main

import (
	"github.com/BurntSushi/toml"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/consumer"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Logger   LoggerConf            `toml:"logger" env-prefix:"LOGGER_"`
	RabbitMQ consumer.RabbitMQConf `toml:"rabbitmq" env-prefix:"RABBITMQ_"`
}

type LoggerConf struct {
	Mod   string `toml:"mod" env:"MOD"`
	Path  string `toml:"path" env:"PATH"`
	Level string `toml:"level" env:"LEVEL"`
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
