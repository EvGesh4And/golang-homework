package producer

// RabbitMQConf defines RabbitMQ producer configuration.
type RabbitMQConf struct {
	URI          string `toml:"uri" env:"URI"`
	Exchange     string `toml:"exchange" env:"EXCHANGE"`
	ExchangeType string `toml:"exchange_type" env:"EXCHANGE_TYPE"`
	RoutingKey   string `toml:"routing_key" env:"ROUTING_KEY"`
	Reliable     bool   `toml:"reliable" env:"RELIABLE"`
}
