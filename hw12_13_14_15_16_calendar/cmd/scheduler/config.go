package main

import (
	"github.com/BurntSushi/toml"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/producer"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/scheduler"
)

type Config struct {
	Logger        LoggerConf                  `toml:"logger"`
	Storage       StorageConf                 `toml:"storage"`
	Notifications scheduler.NotificationsConf `toml:"notifications"`
	RabbitMQ      producer.RabbitMQConf       `toml:"rabbitmq"`
}

type LoggerConf struct {
	Mod   string `toml:"mod"`
	Path  string `toml:"path"`
	Level string `toml:"level"`
}

type StorageConf struct {
	DSN string `toml:"dsn"`
}

func NewConfig() Config {
	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		panic(err)
	}
	return config
}
