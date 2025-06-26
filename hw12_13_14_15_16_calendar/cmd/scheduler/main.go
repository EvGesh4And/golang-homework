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
	cfg, err := NewConfig()
	if err != nil {
		log.Printf("error initializing config: %v", err)
		return
	}
	childLoggers, closer, err := setupLogger(cfg)
	if err != nil {
		log.Fatalf("logger setup error: %v", err)
	}
	if closer != nil {
		defer closer.Close()
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	storage, closer, err := setupStorage(ctx, cfg, childLoggers)
	if err != nil {
		log.Printf("error initializing storage: %v", err)
		return
	}

	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("error closing storage: %s", err)
		} else {
			log.Print("storage sql closed successfully")
		}
	}()

	producer, err := producer.NewRabbitProducer(ctx, cfg.RabbitMQ, childLoggers.scheduler)
	if err != nil {
		return
	}

	scheduler := scheduler.NewScheduler(childLoggers.scheduler, storage, producer, cfg.Notifications)

	scheduler.Start(ctx)
	log.Print("scheduler shutdown complete...")
}
