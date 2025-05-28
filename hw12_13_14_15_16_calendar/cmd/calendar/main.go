package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/app"
	"github.com/EvGesh4And/hw12_13_14_15_calendar/internal/logger"
	internalhttp "github.com/EvGesh4And/hw12_13_14_15_calendar/internal/server/http"
	memorystorage "github.com/EvGesh4And/hw12_13_14_15_calendar/internal/storage/memory"
)

var configFile string

func init() {
	flag.StringVar(&configFile, "config", "configs/config.toml", "Path to configuration file") // /etc/calendar/config.toml
}

func main() {
	flag.Parse()

	if flag.Arg(0) == "version" {
		printVersion()
		return
	}

	config := NewConfig()

	var logg app.Logger
	var storage app.Storage

	logg = logger.New(config.Logger.Level, os.Stdout)

	if config.Storage.mod == "memory" {
		logg.Info("используется in-memory хранилище")
		storage = memorystorage.New(logg)
	}

	calendar := app.New(logg, storage)
	logg.Info("календарь создан")

	server := internalhttp.NewServer(config.HTTP.Host, config.HTTP.Port, logg, calendar)
	logg.Info("http сервер создан")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	// Запускаем сервер в фоне
	go func() {
		addr := fmt.Sprintf("%s:%d", config.HTTP.Host, config.HTTP.Port)
		logg.Info("запускаем HTTP сервер " + addr + "...")

		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Ожидаем сигнала завершения или ошибки от сервера
	select {
	case <-ctx.Done():
		logg.Info("получен сигнал завершения, останавливаем сервер...")
	case err := <-errCh:
		logg.Error("сервер аварийно завершился: " + err.Error())
	}

	// Таймаут на graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		logg.Error("ошибка завершения сервера: " + err.Error())
		os.Exit(1)
	}

	logg.Info("сервер остановлен корректно...")
}
