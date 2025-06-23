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

func NewRabbitConsumer(ctx context.Context, cfg RabbitMQConf, logger *slog.Logger) (*RabbitConsumer, error) {
	c := &RabbitConsumer{
		conn:    nil,
		channel: nil,
		tag:     cfg.ConsumerTag,
		done:    make(chan error, 1),
		logger:  logger,
	}

	const maxAttempts = 5
	const retryDelay = 2 * time.Second

	var err error

	for i := 1; i <= maxAttempts; i++ {
		logger.Info("Попытка подключения к RabbitMQ", slog.String("uri", cfg.URI), slog.Int("attempt", i))
		c.conn, err = amqp.Dial(cfg.URI)
		if err == nil {
			break
		}

		log.Printf("Попытка %d: ошибка подключения к RabbitMQ: %v", i, err)

		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("подключение прервано по контексту: %w", ctx.Err())
		case <-time.After(retryDelay):
			// Пауза перед следующей попыткой
		}
	}

	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к RabbitMQ после %d попыток: %w", maxAttempts, err)
	}

	go func() {
		fmt.Printf("closing: %s", <-c.conn.NotifyClose(make(chan *amqp.Error)))
	}()

	log.Printf("got Connection, getting Channel")
	c.channel, err = c.conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("channel: %w", err)
	}

	log.Printf("got Channel, declaring Exchange (%q)", cfg.Exchange)
	if err = c.channel.ExchangeDeclare(
		cfg.Exchange,     // name of the exchange
		cfg.ExchangeType, // type
		true,             // durable
		false,            // delete when complete
		false,            // internal
		false,            // noWait
		nil,              // arguments
	); err != nil {
		return nil, fmt.Errorf("exchange Declare: %w", err)
	}

	log.Printf("declared Exchange, declaring Queue %q", cfg.Queue)
	queue, err := c.channel.QueueDeclare(
		cfg.Queue, // name of the queue
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // noWait
		nil,       // arguments
	)
	if err != nil {
		return nil, fmt.Errorf("queue Declare: %w", err)
	}

	c.queue = queue

	log.Printf("declared Queue (%q %d messages, %d consumers), binding to Exchange (key %q)",
		queue.Name, queue.Messages, queue.Consumers, cfg.BindingKey)

	if err = c.channel.QueueBind(
		queue.Name,     // name of the queue
		cfg.BindingKey, // bindingKey
		cfg.Exchange,   // sourceExchange
		false,          // noWait
		nil,            // arguments
	); err != nil {
		return nil, fmt.Errorf("queue Bind: %w", err)
	}

	return c, nil
}

func (c *RabbitConsumer) Shutdown() error {
	// will close() the deliveries channel
	if err := c.channel.Cancel(c.tag, true); err != nil {
		return fmt.Errorf("consumer cancel failed: %w", err)
	}

	if err := c.conn.Close(); err != nil {
		return fmt.Errorf("AMQP connection close error: %w", err)
	}

	defer log.Printf("AMQP shutdown OK")

	// wait for handle() to exit
	return <-c.done
}

func (c *RabbitConsumer) Handle(ctx context.Context) error {
	ctx = logger.WithLogMethod(ctx, "Handle")

	c.logger.InfoContext(ctx, "queue bound to Exchange, starting Consume", "consumer_tag", c.tag)
	deliveries, err := c.channel.Consume(
		c.queue.Name, // name
		c.tag,        // consumerTag,
		false,        // noAck
		false,        // exclusive
		false,        // noLocal
		false,        // noWait
		nil,          // arguments
	)
	if err != nil {
		return fmt.Errorf("queue Consume: %w", err)
	}
	for {
		select {
		case <-ctx.Done():
			log.Print("context canceled")
			c.done <- nil
			return ctx.Err()
		case d := <-deliveries:
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
	// log.Printf("handle: deliveries channel closed")
	// c.done <- nil

	// return nil
}
