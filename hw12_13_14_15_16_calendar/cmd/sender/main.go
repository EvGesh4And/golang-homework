package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

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

	log.SetOutput(os.Stdout)
	cfg := NewConfig()
	childLoggers, closer, err := setupLogger(cfg)
	if err != nil {
		return
	}
	if closer != nil {
		defer closer.Close()
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer, err := consumer.NewRabbitConsumer(ctx, cfg.RabbitMQ, childLoggers.sender)
	if err != nil {
		return
	}

	if err := consumer.Handle(ctx); err != nil {
		log.Printf("error during handling: %s", err)
	}

	if err := consumer.Shutdown(); err != nil {
		log.Printf("error during shutdown: %s", err)
		return
	}

	log.Print("sender завершился корректно...")
}
