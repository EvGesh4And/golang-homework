package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/app"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	internalhttp "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/http"
	memorystorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
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

	// Дефолтный logger пишет в ст. вывод.
	log.SetOutput(os.Stdout)

	config := NewConfig()

	var appLogger app.Logger
	var storage app.Storage

	appLogger = logger.New(config.Logger.Level, os.Stdout)

	switch config.Storage.Mod {
	case "memory":
		storage = memorystorage.New()
		log.Print("используется in-memory хранилище")
	case "sql":
		log.Print("инициализация подключения к PostgreSQL...")

		sqlStorage := sqlstorage.New(config.Storage.DSN)
		// Таймаут на установление соединения
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := sqlStorage.Connect(ctx); err != nil {
			log.Printf("ошибка подключения к PostgreSQL: %v", err)
			return
		}

		defer func() {
			if err := sqlStorage.Close(); err != nil {
				log.Print("ошибка закрытия psql подключения", err)
			} else {
				log.Print("psql подключение успешно закрыто")
			}
		}()

		log.Print("выполнение миграций...")
		if err := sqlStorage.Migrate(config.Storage.Migration); err != nil {
			log.Printf("ошибка миграции: %v", err)
			return
		}

		storage = sqlStorage
		log.Print("SQL-хранилище успешно инициализировано и подключено")
	default:
		log.Printf("неизвестный тип хранилища: %v", config.Storage.Mod)
		return
	}

	calendar := app.New(appLogger, storage)
	log.Print("сервис calendar создан")

	server := internalhttp.NewServer(config.HTTP.Host, config.HTTP.Port, appLogger, calendar)
	log.Print("http сервер создан")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)

	// Запускаем сервер в фоне
	go func() {
		addr := fmt.Sprintf("%s:%d", config.HTTP.Host, config.HTTP.Port)
		log.Print("HTTP сервер запускается " + addr + "...")

		if err := server.Start(); err != nil {
			errCh <- err
		}
	}()

	// Ожидаем сигнала завершения или ошибки от сервера
	select {
	case <-ctx.Done():
		log.Print("получен сигнал завершения, останавливаем сервер...")
	case err := <-errCh:
		log.Printf("сервер аварийно остановился: %s", err)
	}

	// Таймаут на graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := server.Stop(shutdownCtx); err != nil {
		log.Printf("[shutdown] ошибка завершения сервера: %s", err)
	} else {
		log.Print("[shutdown] сервер завершился корректно...")
	}
}
