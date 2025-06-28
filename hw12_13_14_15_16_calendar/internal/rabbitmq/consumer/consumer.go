package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/streadway/amqp"
)

// RabbitConsumer consumes events from RabbitMQ.
type RabbitConsumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	tag       string
	queue     amqp.Queue
	logger    *slog.Logger
	cancel    context.CancelFunc
	done      chan error
	reconnect chan struct{}
}

func (c *RabbitConsumer) setLogCompMeth(ctx context.Context, method string) context.Context {
	ctx = logger.WithLogComponent(ctx, "rabbitmq.consumer")
	return logger.WithLogMethod(ctx, method)
}

// NewRabbitConsumer creates and configures a RabbitMQ consumer.
func NewRabbitConsumer(ctx context.Context, cfg RabbitMQConf, lg *slog.Logger) (*RabbitConsumer, error) {
	ctx, cancel := context.WithCancel(ctx)
	c := &RabbitConsumer{
		tag:       cfg.ConsumerTag,
		logger:    lg,
		cancel:    cancel,
		done:      make(chan error),
		reconnect: make(chan struct{}, 1),
	}

	if err := c.connectWithRetry(ctx, cfg.URI); err != nil {
		return nil, err
	}

	if err := c.initChannel(ctx); err != nil {
		return nil, err
	}

	if err := c.declareExchangeQueueBind(ctx, cfg); err != nil {
		return nil, err
	}

	c.reconnect <- struct{}{}
	go c.startReconnectLoop(ctx, cfg)

	return c, nil
}

func (c *RabbitConsumer) connectWithRetry(ctx context.Context, uri string) error {
	ctx = c.setLogCompMeth(ctx, "connectWithRetry")
	const (
		maxAttempts = 5
		retryDelay  = 2 * time.Second
	)

	var err error
	for i := 1; i <= maxAttempts; i++ {
		c.logger.DebugContext(ctx, "try connecting to RabbitMQ", slog.String("uri", uri), slog.Int("attempt", i))

		c.conn, err = amqp.Dial(uri)
		if err == nil {
			break
		}

		c.logger.WarnContext(ctx, "failed to connect to RabbitMQ", "error", err)

		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "connection cancelled", "error", ctx.Err())
			return logger.WrapError(ctx, fmt.Errorf("connection cancelled: %w", ctx.Err()))
		case <-time.After(retryDelay):
			// Pause before the next attempt
		}
	}
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w",
			maxAttempts, err))
	}

	c.logger.InfoContext(ctx, "connection established")

	// отслеживаем закрытие подключения брокером
	go func() {
		c.logger.WarnContext(ctx, "connection closed", "error", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	return nil
}

func (c *RabbitConsumer) initChannel(ctx context.Context) error {
	ctx = c.setLogCompMeth(ctx, "initChannel")
	c.logger.DebugContext(ctx, "trying to initialize channel")
	var err error
	c.channel, err = c.conn.Channel()
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("channel: %w", err))
	}
	c.logger.InfoContext(ctx, "channel initialized")
	return nil
}

func (c *RabbitConsumer) declareExchangeQueueBind(ctx context.Context, cfg RabbitMQConf) error {
	ctx = c.setLogCompMeth(ctx, "declareExchangeQueueBind")

	c.logger.DebugContext(ctx, "try declaring exchange", "type", cfg.ExchangeType, "name", cfg.Exchange)

	if err := c.channel.ExchangeDeclare(
		cfg.Exchange,
		cfg.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("exchange declare: %w", err))
	}

	c.logger.InfoContext(ctx, "exchange declared")

	c.logger.DebugContext(ctx, "try declaring queue", "name", cfg.Queue)

	q, err := c.channel.QueueDeclare(
		cfg.Queue, true, false, false, false, nil,
	)
	if err != nil {
		return logger.WrapError(ctx, fmt.Errorf("queue declare: %w", err))
	}
	c.queue = q

	c.logger.InfoContext(ctx, "queue declared")

	c.logger.DebugContext(ctx, "try binding queue", "name", q.Name, "binding_key",
		cfg.BindingKey, "exchange", cfg.Exchange)

	if err := c.channel.QueueBind(
		q.Name, cfg.BindingKey, cfg.Exchange, false, nil,
	); err != nil {
		return logger.WrapError(ctx, fmt.Errorf("queue bind: %w", err))
	}

	c.logger.InfoContext(ctx, "queue bound")

	return nil
}

// Handle starts consuming messages from RabbitMQ.
func (c *RabbitConsumer) Handle(ctx context.Context) error {
	defer close(c.done)

	ctx = c.setLogCompMeth(ctx, "Handle")

outer:
	for {
		select {
		case <-ctx.Done():
			c.logger.InfoContext(ctx, "context cancelled")
			return logger.WrapError(ctx, fmt.Errorf("context cancelled: %w", ctx.Err()))
		case <-c.reconnect:
		}

		c.logger.DebugContext(ctx, "try consuming", "consumer_tag", c.tag)

		deliveries, err := c.channel.Consume(
			c.queue.Name,
			c.tag,
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			return logger.WrapError(ctx, fmt.Errorf("queue consume: %w", err))
		}

		c.logger.InfoContext(ctx, "messages are being consumed")

		for {
			select {
			case <-ctx.Done():
				c.logger.InfoContext(ctx, "context cancelled")
				return logger.WrapError(ctx, fmt.Errorf("context cancelled: %w", ctx.Err()))
			case d, ok := <-deliveries:
				if !ok {
					c.logger.InfoContext(ctx, "deliveries channel closed")
					continue outer
				}

				c.logger.DebugContext(ctx,
					"received delivery",
					"size", len(d.Body),
					"tag", d.DeliveryTag,
					"body", string(d.Body),
				)

				c.logger.InfoContext(ctx, "message delivered", "delivery_tag", d.DeliveryTag)

				c.logger.DebugContext(ctx, "try unmarshalling notification")
				var notification storage.Notification
				if err := json.Unmarshal(d.Body, &notification); err != nil {
					return logger.WrapError(ctx, fmt.Errorf("error during unmarshal: %w", err))
				}
				c.logger.InfoContext(ctx, "notification unmarshalled", "notification", notification)

				if err := c.ackDelivery(ctx, d); err != nil {
					return err
				}

				c.logger.InfoContext(ctx, "notification event", "notification", notification)
			}
		}
	}
}

// Shutdown gracefully closes consumer resources.
func (c *RabbitConsumer) Shutdown() error {
	var errs []error

	if err := c.channel.Cancel(c.tag, true); err != nil {
		errs = append(errs, fmt.Errorf("RabbitConsumer.Shutdown: consumer cancel failed: %w", err))
	}

	if err := c.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("RabbitConsumer.Shutdown: AMQP connection close error: %w", err))
	}

	c.cancel()
	// Ожидаем завершения Handle
	<-c.done

	return errors.Join(errs...)
}

func (c *RabbitConsumer) ackDelivery(ctx context.Context, d amqp.Delivery) error {
	if err := d.Ack(false); err != nil {
		c.logger.ErrorContext(ctx, "failed to acknowledge message", slog.String("error", err.Error()))
		return logger.WrapError(ctx, fmt.Errorf("ack: %w", err))
	}
	return nil
}

func (c *RabbitConsumer) startReconnectLoop(ctx context.Context, cfg RabbitMQConf) {
	notify := c.conn.NotifyClose(make(chan *amqp.Error))
	for {
		select {
		case <-ctx.Done():
			return
		case errNotify := <-notify:
			if errNotify != nil {
				c.logger.WarnContext(ctx, "connection closed", slog.String("error", errNotify.Error()))
			}

			for ctx.Err() == nil {
				if err := c.connectWithRetry(ctx, cfg.URI); err != nil {
					continue
				}
				if err := c.initChannel(ctx); err != nil {
					continue
				}
				if err := c.declareExchangeQueueBind(ctx, cfg); err != nil {
					continue
				}

				notify = c.conn.NotifyClose(make(chan *amqp.Error))

				select {
				case c.reconnect <- struct{}{}:
				default:
				}

				break
			}
		}
	}
}
