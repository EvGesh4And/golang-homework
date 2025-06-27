package producer

import (
	"context"
	"errors"
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
	reliable   bool
	confirms   chan amqp.Confirmation
	logger     *slog.Logger
}

func NewRabbitProducer(ctx context.Context, cfg RabbitMQConf, logger *slog.Logger) (*RabbitProducer, error) {
	p := &RabbitProducer{
		exchange:   cfg.Exchange,
		routingKey: cfg.RoutingKey,
		reliable:   cfg.Reliable,
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
		p.logger.DebugContext(ctx, "try connecting to RabbitMQ", slog.String("uri", uri),
			slog.Int("attempt", i))

		p.conn, err = amqp.Dial(uri)
		if err == nil {
			break
		}

		p.logger.WarnContext(ctx, "failed to connect to RabbitMQ", slog.Int("attempt", i),
			slog.String("error", err.Error()))

		select {
		case <-ctx.Done():
			p.logger.InfoContext(ctx, "connection cancelled", "error", ctx.Err())
			return fmt.Errorf("RabbitProducer.connectWithRetry: connection cancelled: %w", ctx.Err())
		case <-time.After(retryDelay):
			// Pause before the next attempt
		}
	}

	if err != nil {
		return fmt.Errorf("RabbitProducer.connectWithRetry: failed to connect to RabbitMQ after %d attempts: %w",
			maxAttempts, err)
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

	// Включаем publisher confirms
	if p.reliable {
		p.logger.InfoContext(ctx, "enabling publishing confirms")

		if err := p.channel.Confirm(false); err != nil {
			return fmt.Errorf("RabbitProducer.initChannel: could not enable confirms: %w", err)
		}

		p.confirms = p.channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	}

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

	if err := p.channel.Publish(
		p.exchange,
		p.routingKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     "text/plain",
			ContentEncoding: "",
			Body:            []byte(body),
			DeliveryMode:    amqp.Transient,
			Priority:        0,
		},
	); err != nil {
		return fmt.Errorf("RabbitProducer.Publish: failed to publish message: %w", err)
	}

	p.logger.InfoContext(ctx, "RabbitProducer.Publish: message published", "body", body)

	// Ожидание подтверждения (если reliable)
	if p.reliable {
		select {
		case confirm := <-p.confirms:
			if confirm.Ack {
				p.logger.InfoContext(ctx, "message delivery confirmed", slog.Uint64("deliveryTag", confirm.DeliveryTag))
			} else {
				p.logger.ErrorContext(ctx, "message delivery NOT confirmed", slog.Uint64("deliveryTag", confirm.DeliveryTag))
				return fmt.Errorf("RabbitProducer.Publish: message not acknowledged by broker")
			}
		case <-time.After(5 * time.Second):
			return fmt.Errorf("RabbitProducer.Publish: timeout waiting for confirmation")
		}
	}

	return nil
}

func (p *RabbitProducer) Shutdown() error {
	var errs []error

	if err := p.channel.Close(); err != nil {
		errs = append(errs, fmt.Errorf("RabbitProducer.Shutdown: channel close failed: %w", err))
	}

	if err := p.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("RabbitProducer.Shutdown: AMQP connection close error: %w", err))
	}

	return errors.Join(errs...)
}
