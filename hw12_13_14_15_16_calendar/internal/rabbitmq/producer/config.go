package producer

type RabbitMQConf struct {
	URI          string `toml:"uri"`
	Exchange     string `toml:"exchange"`
	ExchangeType string `toml:"exchange_type"`
	RoutingKey   string `toml:"routing_key"`
	Reliable     bool   `toml:"reliable"`
}
