package consumer

type RabbitMQConf struct {
	URI          string `toml:"uri"`
	Exchange     string `toml:"exchange"`
	ExchangeType string `toml:"exchange_type"`
	Queue        string `toml:"queue"`
	BindingKey   string `toml:"binding_key"`
	ConsumerTag  string `toml:"consumer_tag"`
}
