package producer

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/streadway/amqp"
)

type RabbitProducer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	exchange   string
	routingKey string
	logger     *slog.Logger
}

func NewRabbitProducer(ctx context.Context, cfg RabbitMQConf, logger *slog.Logger) (*RabbitProducer, error) {
	const maxAttempts = 5
	const retryDelay = 2 * time.Second

	var conn *amqp.Connection
	var err error

	for i := 1; i <= maxAttempts; i++ {
		logger.Info("attempting to connect to RabbitMQ", slog.String("uri", cfg.URI), slog.Int("attempt", i))
		conn, err = amqp.Dial(cfg.URI)
		if err == nil {
			break
		}

		log.Printf("Attempt %d: failed to connect to RabbitMQ: %v", i, err)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("connection aborted by context: %w", ctx.Err())
		case <-time.After(retryDelay):
			// Пауза перед следующей попыткой
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxAttempts, err)
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
		logger:     logger,
	}, nil
}

func (p *RabbitProducer) Publish(ctx context.Context, body string) error {
	ctx = logger.WithLogMethod(ctx, "Publish")
	p.logger.InfoContext(ctx, "publishing message", "body", body)

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
