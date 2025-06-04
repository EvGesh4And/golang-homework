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
	flag.StringVar(&configFile, "config", "configs/config.toml", "Path to configuration file")
}

func main() {
	flag.Parse()
	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	log.SetOutput(os.Stdout)
	cfg := NewConfig()
	childLoggers := setupLogger(cfg)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	storage := setupStorage(cfg, childLoggers)
	calendar := app.New(childLoggers.app, storage)

	wg := sync.WaitGroup{}
	startHTTPServer(&wg, ctx, cfg, childLoggers.http, calendar)
	startGRPCServer(&wg, ctx, cfg, childLoggers.grpc, calendar)
	wg.Wait()
}
