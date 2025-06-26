package main

import (
	"github.com/BurntSushi/toml"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/producer"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/scheduler"
	"github.com/caarlos0/env/v10"
)

type Config struct {
	Logger        LoggerConf                  `toml:"logger" env-prefix:"LOGGER_"`
	Storage       StorageConf                 `toml:"storage" env-prefix:"STORAGE_"`
	Notifications scheduler.NotificationsConf `toml:"notifications" env-prefix:"NOTIFICATIONS_"`
	RabbitMQ      producer.RabbitMQConf       `toml:"rabbitmq" env-prefix:"RABBITMQ_"`
}

type LoggerConf struct {
	Mod   string `toml:"mod" env:"MOD"`
	Path  string `toml:"path" env:"PATH"`
	Level string `toml:"level" env:"LEVEL"`
}

type StorageConf struct {
	DSN string `toml:"dsn" env:"DSN"`
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
