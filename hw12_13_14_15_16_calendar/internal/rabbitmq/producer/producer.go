package producer

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitProducer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	routingKey string
}

func NewRabbitProducer(cfg RabbitMQConf) (*RabbitProducer, error) {
	log.Printf("dialing %q", cfg.URI)
	conn, err := amqp.Dial(cfg.URI)
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("channel: %w", err)
	}
	log.Printf("got Channel, declaring %q Exchange (%q)", cfg.ExchangeType, cfg.Exchange)

	if err := ch.ExchangeDeclare(
		cfg.Exchange,     // name
		cfg.ExchangeType, // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // noWait
		nil,              // arguments
	); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("exchange declare: %w", err)
	}

	return &RabbitProducer{
		conn:       conn,
		channel:    ch,
		exchange:   cfg.Exchange,
		routingKey: cfg.RoutingKey,
	}, nil
}

func (p *RabbitProducer) Publish(body string) error {
	return p.channel.Publish(
		p.exchange,   // publish to an exchange
		p.routingKey, // routing to 0 or more queues
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient, // 1=non-persistent, 2=persistent
			Priority:        0,              // 0-9
			// a bunch of application/implementation-specific fields
		},
	)
}

func (p *RabbitProducer) Close() {
	_ = p.channel.Close()
	_ = p.conn.Close()
}
