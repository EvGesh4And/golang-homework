package main

import (
	"time"

	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger        LoggerConf        `toml:"logger"`
	Storage       StorageConf       `toml:"storage"`
	Notifications NotificationsConf `toml:"notifications"`
}

type LoggerConf struct {
	Mod   string `toml:"mod"`
	Path  string `toml:"path"`
	Level string `toml:"level"`
}

type StorageConf struct {
	DSN string `toml:"dsn"`
}

type NotificationsConf struct {
	Tick time.Duration `toml:"tick"`
}

func NewConfig() Config {
	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		panic(err)
	}
	return config
}
