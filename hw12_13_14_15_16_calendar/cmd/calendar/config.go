package main

import "github.com/BurntSushi/toml"

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf  `toml:"logger"`
	Storage StorageConf `toml:"storage"`
	HTTP    HTTPConf    `toml:"http"`
	GRPC    GRPCConf    `toml:"grpc"`
}

type LoggerConf struct {
	Mod   string `toml:"mod"`
	Path  string `toml:"path"`
	Level string `toml:"level"`
}

type StorageConf struct {
	Mod       string `toml:"mod"`
	DSN       string `toml:"dsn"`
	Migration string `toml:"migration"`
}

type HTTPConf struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

type GRPCConf struct {
	Host string `toml:"host"`
	Port int    `toml:"port"`
}

func NewConfig() Config {
	var config Config
	if _, err := toml.DecodeFile(configFile, &config); err != nil {
		panic(err)
	}
	return config
}
