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
	p := &RabbitProducer{
		exchange:   cfg.Exchange,
		routingKey: cfg.RoutingKey,
		logger:     logger,
	}

	if err := p.connectWithRetry(ctx, cfg.URI); err != nil {
		return nil, err
	}

	if err := p.initChannel(); err != nil {
		return nil, err
	}

	if err := p.setupExchange(cfg); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *RabbitProducer) connectWithRetry(ctx context.Context, uri string) error {
	const (
		maxAttempts = 5
		retryDelay  = 2 * time.Second
	)

	var err error
	for i := 1; i <= maxAttempts; i++ {
		p.logger.Info("Попытка подключения к RabbitMQ", slog.String("uri", uri), slog.Int("attempt", i))

		p.conn, err = amqp.Dial(uri)
		if err == nil {
			break
		}

		log.Printf("Попытка %d: ошибка подключения к RabbitMQ: %v", i, err)

		select {
		case <-ctx.Done():
			return fmt.Errorf("подключение прервано по контексту: %w", ctx.Err())
		case <-time.After(retryDelay):
			// Пауза перед следующей попыткой
		}
	}

	if err != nil {
		return fmt.Errorf("не удалось подключиться к RabbitMQ после %d попыток: %w", maxAttempts, err)
	}

	return nil
}

func (p *RabbitProducer) initChannel() error {
	var err error
	p.channel, err = p.conn.Channel()
	if err != nil {
		return fmt.Errorf("channel: %w", err)
	}
	return nil
}

func (p *RabbitProducer) setupExchange(cfg RabbitMQConf) error {
	log.Printf("got Channel, declaring %q Exchange (%q)", cfg.ExchangeType, cfg.Exchange)

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
		return fmt.Errorf("exchange declare: %w", err)
	}

	return nil
}

func (p *RabbitProducer) Publish(ctx context.Context, body string) error {
	ctx = logger.WithLogMethod(ctx, "Publish")
	p.logger.InfoContext(ctx, "публикация сообщения", "body", body)

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
		p.logger.ErrorContext(ctx, "ошибка публикации сообщения", "error", err)
	}

	return err
}

func (p *RabbitProducer) Shutdown() error {
	// останавливаем получение новых сообщений
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("consumer cancel failed: %w", err)
	}

	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %w", err)
	}
	return nil
}
