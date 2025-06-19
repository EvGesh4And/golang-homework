package main

import (
	"github.com/BurntSushi/toml"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/consumer"
)

type Config struct {
	Logger   LoggerConf            `toml:"logger"`
	RabbitMQ consumer.RabbitMQConf `toml:"rabbitmq"`
}

type LoggerConf struct {
	Mod   string `toml:"mod"`
	Path  string `toml:"path"`
	Level string `toml:"level"`
}

func NewConfig() Config {
	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		panic(err)
	}
	return config
}
