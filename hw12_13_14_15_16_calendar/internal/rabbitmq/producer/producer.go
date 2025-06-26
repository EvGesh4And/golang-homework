package producer

import (
	"context"
	"fmt"
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
	p := &RabbitProducer{
		exchange:   cfg.Exchange,
		routingKey: cfg.RoutingKey,
		logger:     logger,
	}

	if err := p.connectWithRetry(ctx, cfg.URI); err != nil {
		return nil, err
	}

	if err := p.initChannel(ctx); err != nil {
		return nil, err
	}

	if err := p.setupExchange(ctx, cfg); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *RabbitProducer) connectWithRetry(ctx context.Context, uri string) error {
	ctx = logger.WithLogMethod(ctx, "connectWithRetry")
	const (
		maxAttempts = 5
		retryDelay  = 2 * time.Second
	)

	var err error
	for i := 1; i <= maxAttempts; i++ {
		p.logger.DebugContext(ctx, "try connecting to RabbitMQ", slog.String("uri", uri), slog.Int("attempt", i))

		p.conn, err = amqp.Dial(uri)
		if err == nil {
			break
		}

		p.logger.WarnContext(ctx, "failed to connect to RabbitMQ", slog.Int("attempt", i), slog.String("error", err.Error()))

		select {
		case <-ctx.Done():
			p.logger.InfoContext(ctx, "connection cancelled", "error", ctx.Err())
			return fmt.Errorf("RabbitProducer.connectWithRetry: connection cancelled: %w", ctx.Err())
		case <-time.After(retryDelay):
			// Pause before the next attempt
		}
	}

	if err != nil {
		return fmt.Errorf("RabbitProducer.connectWithRetry: failed to connect to RabbitMQ after %d attempts: %w", maxAttempts, err)
	}

	p.logger.InfoContext(ctx, "connection established")

	// отслеживаем закрытие подключения брокером
	go func() {
		p.logger.WarnContext(ctx, "connection closed", "error", <-p.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	return nil
}

func (p *RabbitProducer) initChannel(ctx context.Context) error {
	ctx = logger.WithLogMethod(ctx, "initChannel")
	p.logger.DebugContext(ctx, "trying to initialize channel")
	var err error
	p.channel, err = p.conn.Channel()
	if err != nil {
		return fmt.Errorf("RabbitProducer.initChannel: channel: %w", err)
	}
	p.logger.InfoContext(ctx, "channel initialized")
	return nil
}

func (p *RabbitProducer) setupExchange(ctx context.Context, cfg RabbitMQConf) error {
	ctx = logger.WithLogMethod(ctx, "setupExchange")

	p.logger.DebugContext(ctx, "try declaring exchange", "type", cfg.ExchangeType, "name", cfg.Exchange)

	if err := p.channel.ExchangeDeclare(
		cfg.Exchange,     // name
		cfg.ExchangeType, // type
		true,             // durable
		false,            // auto-deleted
		false,            // internal
		false,            // noWait
		nil,              // arguments
	); err != nil {
		p.channel.Close()
		return fmt.Errorf("RabbitProducer.setupExchange: exchange declare: %w", err)
	}

	p.logger.InfoContext(ctx, "exchange declared")

	return nil
}

func (p *RabbitProducer) Publish(ctx context.Context, body string) error {
	ctx = logger.WithLogMethod(ctx, "Publish")
	p.logger.DebugContext(ctx, "publishing message", "body", body)

	err := p.channel.Publish(
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
	if err != nil {
		return fmt.Errorf("RabbitProducer.Publish: failed to publish message: %w", err)
	}
	p.logger.InfoContext(ctx, "message published", "body", body)

	return nil
}

func (p *RabbitProducer) Shutdown() error {
	// останавливаем получение новых сообщений
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("RabbitProducer.Shutdown: consumer cancel failed: %w", err)
	}

	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("RabbitProducer.Shutdown: AMQP connection close error: %w", err)
	}
	return nil
}
