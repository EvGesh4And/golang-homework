package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/streadway/amqp"
)

type RabbitConsumer struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	tag     string
	done    chan error
	queue   amqp.Queue
	logger  *slog.Logger
}

func NewRabbitConsumer(ctx context.Context, cfg RabbitMQConf, lg *slog.Logger) (*RabbitConsumer, error) {
	c := &RabbitConsumer{
		tag:    cfg.ConsumerTag,
		done:   make(chan error, 1),
		logger: lg,
	}

	const (
		maxAttempts = 5
		retryDelay  = 2 * time.Second
	)

	var err error
	for i := 1; i <= maxAttempts; i++ {
		lg.Info("connecting to RabbitMQ",
			slog.String("uri", cfg.URI),
			slog.Int("attempt", i),
		)

		c.conn, err = amqp.Dial(cfg.URI)
		if err == nil {
			break
		}

		lg.Error("RabbitMQ dial failed", "error", err)
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("dial cancelled: %w", ctx.Err())
		case <-time.After(retryDelay):
		}
	}
	if err != nil {
		return nil, fmt.Errorf("unable to connect after %d attempts: %w", maxAttempts, err)
	}

	// отслеживаем закрытие подключения брокером
	go func() {
		c.logger.Warn("connection closed", "error", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	// канал
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %w", err)
	}

	// exchange + queue + bind
	if err = c.channel.ExchangeDeclare(
		cfg.Exchange, cfg.ExchangeType,
		true, false, false, false, nil,
	); err != nil {
		return nil, fmt.Errorf("exchange declare: %w", err)
	}

	q, err := c.channel.QueueDeclare(
		cfg.Queue, true, false, false, false, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("queue declare: %w", err)
	}
	c.queue = q

	if err = c.channel.QueueBind(
		q.Name, cfg.BindingKey, cfg.Exchange, false, nil,
	); err != nil {
		return nil, fmt.Errorf("queue bind: %w", err)
	}

	return c, nil
}

func (c *RabbitConsumer) Handle(ctx context.Context) error {
	ctx = logger.WithLogMethod(ctx, "Handle")
	c.logger.InfoContext(ctx, "start consuming", "consumer_tag", c.tag)

	deliveries, err := c.channel.Consume(
		c.queue.Name, c.tag,
		false, /* auto-ack — нет */
		false, false, false, nil,
	)
	if err != nil {
		return fmt.Errorf("queue consume: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			// контекст отменён, выходим
			c.logger.InfoContext(ctx, "context cancelled")
			close(c.done)
			return ctx.Err()

		case d, ok := <-deliveries:
			if !ok {
				// канал закрыт брокером
				c.logger.InfoContext(ctx, "deliveries channel closed")
				close(c.done)
				return nil
			}

			log.Printf(
				"got %dB delivery: [%v] %q",
				len(d.Body),
				d.DeliveryTag,
				d.Body,
			)
			c.logger.InfoContext(ctx, "message delivered", "delivery_tag", d.DeliveryTag)
			d.Ack(false)

			var notification storage.Notification
			if err := json.Unmarshal(d.Body, &notification); err != nil {
				c.logger.ErrorContext(ctx, "error during unmarshal", "error", err)
				return err
			}

			c.logger.InfoContext(ctx, "notification event", "notification", notification)
		}
	}
}

func (c *RabbitConsumer) Shutdown() error {
	// останавливаем получение новых сообщений
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("consumer cancel failed: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %w", err)
	}

	<-c.done // ждём завершения Handle
	return nil
}
