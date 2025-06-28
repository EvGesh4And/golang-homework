package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/rabbitmq/producer"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/scheduler"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/scheduler_config.toml", "Path to configuration file")
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

	storage, closer, err := setupStorage(ctx, cfg, childLoggers)
	if err != nil {
		return
	}

	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("error closing sql storage: %s", err)
		} else {
			log.Print("sql storage closed successfully")
		}
	}()

	producer, err := producer.NewRabbitProducer(ctx, cfg.RabbitMQ, childLoggers.scheduler)
	if err != nil {
		return
	}

	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("error closing pubsub: %s", err)
		} else {
			log.Print("pubsub closed successfully")
		}
	}()

	scheduler := scheduler.NewScheduler(childLoggers.scheduler, storage, producer, cfg.Notifications)

	scheduler.Start(ctx)
	log.Print("scheduler finished gracefully...")
}
