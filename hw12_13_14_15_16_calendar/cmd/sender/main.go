package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/consumer"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/sender_config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	// ---------- настройка логирования ----------
	cfg := NewConfig()
	child, closer, err := setupLogger(cfg)
	if err != nil {
		log.Fatalf("logger setup error: %v", err)
	}
	if closer != nil {
		defer closer.Close()
	}
	log.SetOutput(os.Stdout)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cons, err := consumer.NewRabbitConsumer(ctx, cfg.RabbitMQ, child.sender)
	if err != nil {
		log.Printf("cannot create consumer: %v", err)
		return
	}

	// ---------- запуск consumer в отдельной горутине ----------
	go func() {
		if err := cons.Handle(ctx); err != nil && !errors.Is(err, context.Canceled) {
			log.Printf("consumer error: %v", err)
		}
	}()

	<-ctx.Done() // ждём SIGINT/SIGTERM

	// ---------- даём немного времени на корректное закрытие ----------
	shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cons.Shutdown(); err != nil {
		log.Printf("error during shutdown: %v", err)
	}

	log.Print("sender завершился корректно")
	_ = shCtx // т.к. Shutdown блокируется лишь до выполнения <-c.done, тайм-аут здесь символический
}
