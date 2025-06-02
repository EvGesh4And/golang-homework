package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/app"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/gRPC/pb"
	grpcserver "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/gRPC/server"
	internalhttp "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/http"
	memorystorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
	"google.golang.org/grpc"
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

	serverHTTP := internalhttp.NewServerHTTP(config.HTTP.Host, config.HTTP.Port, appLogger, calendar)
	log.Print("http сервер создан")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Канал ошибок от HTTP сервера
	errChHTTP := make(chan error, 1)

	wg := sync.WaitGroup{}

	// Запускаем сервер в фоне
	go func() {
		addr := fmt.Sprintf("%s:%d", config.HTTP.Host, config.HTTP.Port)
		log.Print("HTTP сервер запускается " + addr + "...")

		if err := serverHTTP.Start(); err != nil {
			errChHTTP <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		// Ожидаем сигнала завершения или ошибки от сервера
		select {
		case <-ctx.Done():
			log.Print("получен сигнал завершения, останавливаем сервер...")
		case err := <-errChHTTP:
			log.Printf("сервер аварийно остановился: %s", err)
		}

		// Таймаут на graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := serverHTTP.Stop(shutdownCtx); err != nil {
			log.Printf("[shutdown] ошибка завершения сервера HTTP: %s", err)
		} else {
			log.Print("[shutdown] HTTP сервер завершился корректно...")
		}
	}()

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", config.GRPC.Host, config.GRPC.Port))
	if err != nil {
		log.Printf("gRPC не удалось слушать порт %v: %v", fmt.Sprintf("%s:%d", config.GRPC.Host, config.GRPC.Port), err)
		return
	}

	serverGRPC := grpcserver.NewServerGRPC(lis, calendar)

	grpcSrv := grpc.NewServer()
	pb.RegisterCalendarServer(grpcSrv, serverGRPC)

	// Канал ошибок от gRPC сервера
	errChGRPC := make(chan error, 1)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Print("gRPC сервер запускается " + lis.Addr().String() + "...")
		if err := grpcSrv.Serve(lis); err != nil {
			errChGRPC <- err
		}
	}()

	wg.Add(1)
	// Обработка завершения
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			log.Print("получен сигнал завершения, останавливаем gRPC сервер...")
		case err := <-errChGRPC:
			log.Printf("gRPC сервер аварийно остановился: %s", err)
			return
		}

		// Таймаут на graceful shutdown
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop() // корректное завершение
			close(done)
		}()

		select {
		case <-done:
			log.Print("[shutdown] gRPC сервер завершился корректно...")
		case <-shutdownCtx.Done():
			log.Print("[shutdown] таймаут graceful shutdown gRPC, вызываем Stop()")
			grpcSrv.Stop() // экстренная остановка
		}
	}()

	wg.Wait()
}
