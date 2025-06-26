package consumer

type RabbitMQConf struct {
	URI          string `toml:"uri" env:"URI"`
	Exchange     string `toml:"exchange" env:"EXCHANGE"`
	ExchangeType string `toml:"exchange_type" env:"EXCHANGE_TYPE"`
	Queue        string `toml:"queue" env:"QUEUE"`
	BindingKey   string `toml:"binding_key" env:"BINDING_KEY"`
	ConsumerTag  string `toml:"consumer_tag" env:"CONSUMER_TAG"`
}
