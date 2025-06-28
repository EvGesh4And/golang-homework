package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/app"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"golang.org/x/sync/errgroup"
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
	cfg, err := NewConfig()
	if err != nil {
		log.Printf("error initializing config: %v", err)
		return
	}
	lg, closer, err := logger.NewLogger(cfg.Logger)
	if err != nil {
		log.Printf("error initializing logger: %v", err)
		return
	}
	if closer != nil {
		defer closer.Close()
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	storage, closer, err := setupStorage(ctx, cfg, lg)
	if err != nil {
		log.Printf("error initializing storage: %v", err)
		return
	}

	defer func() {
		if err := closer.Close(); err != nil {
			log.Printf("error closing storage %s: %s", cfg.Storage.Mod, err)
		} else {
			log.Printf("storage %s successfully closed", cfg.Storage.Mod)
		}
	}()

	calendar := app.New(lg, storage)

	g, ctx := errgroup.WithContext(ctx)

	startHTTPServer(ctx, g, cfg, lg, calendar)
	startGRPCServer(ctx, g, cfg, lg, calendar)

	if err := g.Wait(); err != nil {
		log.Printf("service stopped with error: %v", err)
		return
	}
}
