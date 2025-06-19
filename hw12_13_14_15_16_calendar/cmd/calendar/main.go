package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/app"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/calendar_config.toml", "Path to configuration file")
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
			log.Printf("ошибка закрытия хранилища %s: %s", cfg.Storage.Mod, err)
		} else {
			log.Printf("хранилище %s успешно закрыто", cfg.Storage.Mod)
		}
	}()

	calendar := app.New(childLoggers.app, storage)

	wg := sync.WaitGroup{}
	startHTTPServer(ctx, &wg, cfg, childLoggers.http, calendar)
	startGRPCServer(ctx, &wg, cfg, childLoggers.grpc, calendar)
	wg.Wait()
}
