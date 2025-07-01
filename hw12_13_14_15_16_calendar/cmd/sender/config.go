package main

import (
	"github.com/BurntSushi/toml"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/consumer"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Logger   logger.Config         `toml:"logger" env-prefix:"LOGGER_"`
	RabbitMQ consumer.RabbitMQConf `toml:"rabbitmq" env-prefix:"RABBITMQ_"`
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
