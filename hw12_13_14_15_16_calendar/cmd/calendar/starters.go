package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	pb "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/api"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/app"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	grpcserver "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/grpc"
	internalhttp "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/server/http"
	memorystorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/memory"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
	"google.golang.org/grpc"
)

type ChildLoggers struct {
	app        *slog.Logger
	storageMem *slog.Logger
	storageSQL *slog.Logger
	http       *slog.Logger
	grpc       *slog.Logger
}

func setupLogger(cfg Config) (*ChildLoggers, io.Closer, error) {
	var err error
	var logFile *os.File
	var globalLogger *slog.Logger

	switch cfg.Logger.Mod {
	case "console":
		globalLogger = logger.New(cfg.Logger.Level, os.Stdout)
	case "file":
		filePath := cfg.Logger.Path
		if filePath == "" {
			filePath = "calendar.log" // путь по умолчанию, если не задан
		}

		logFile, err = os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			log.Printf("не удалось открыть лог-файл %s: %s", filePath, err.Error())
			return nil, nil, err
		}
		globalLogger = logger.New(cfg.Logger.Level, logFile)
	default:
		log.Printf("неизвестный режим логгера: %s, используется консоль", cfg.Logger.Mod)
		globalLogger = logger.New(cfg.Logger.Level, os.Stdout)
	}

	childLoggers := &ChildLoggers{
		app:        globalLogger.With("component", "app"),
		storageMem: globalLogger.With("component", "storage", "type", "inmemory"),
		storageSQL: globalLogger.With("component", "storage", "type", "sql"),
		http:       globalLogger.With("component", "http"),
		grpc:       globalLogger.With("component", "grpc"),
	}

	return childLoggers, logFile, nil
}

func setupStorage(ctx context.Context, cfg Config, childLoggers *ChildLoggers) (app.Storage, io.Closer, error) {
	logStorageMem := childLoggers.storageMem
	logStorageSQL := childLoggers.storageSQL

	switch cfg.Storage.Mod {
	case "memory":
		log.Print("используется in-memory хранилище")
		return memorystorage.New(logStorageMem), nil, nil

	case "sql":
		log.Print("инициализация подключения к PostgreSQL...")

		sqlStorage := sqlstorage.New(logStorageSQL, cfg.Storage.DSN)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := sqlStorage.Connect(ctx); err != nil {
			log.Printf("ошибка подключения к PostgreSQL: %v", err)
			return nil, nil, err
		}

		log.Print("выполнение миграций...")
		if err := sqlStorage.Migrate(cfg.Storage.Migration); err != nil {
			log.Print(err)
			return nil, nil, err
		}

		log.Print("SQL-хранилище успешно инициализировано и подключено")
		return sqlStorage, sqlStorage, nil

	default:
		log.Printf("неизвестный тип хранилища: %v", cfg.Storage.Mod)
		return nil, nil, fmt.Errorf("неизвестный тип хранилища: %v", cfg.Storage.Mod)
	}
}

func startHTTPServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	cfg Config,
	logHTTP *slog.Logger,
	calendar *app.App,
) {
	serverHTTP := internalhttp.NewServerHTTP(cfg.HTTP.Host, cfg.HTTP.Port, logHTTP, calendar)
	log.Print("http сервер создан")

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	errCh := make(chan error, 1)

	go func() {
		log.Print("HTTP сервер запускается " + addr + "...")
		if err := serverHTTP.Start(); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			log.Print("получен сигнал завершения, останавливаем HTTP сервер...")
		case err := <-errCh:
			log.Printf("HTTP сервер аварийно остановился: %s", err)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := serverHTTP.Stop(shutdownCtx); err != nil {
			log.Printf("[shutdown] ошибка завершения сервера HTTP: %s", err)
		} else {
			log.Print("[shutdown] HTTP сервер завершился корректно...")
		}
	}()
}

func startGRPCServer(
	ctx context.Context,
	wg *sync.WaitGroup,
	cfg Config,
	logGRPC *slog.Logger,
	calendar *app.App,
) {
	addr := fmt.Sprintf("%s:%d", cfg.GRPC.Host, cfg.GRPC.Port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("gRPC не удалось слушать порт %v: %v", addr, err)
		os.Exit(1)
	}

	serverGRPC := grpcserver.NewServerGRPC(logGRPC, lis, calendar)
	grpcSrv := grpc.NewServer()
	pb.RegisterCalendarServer(grpcSrv, serverGRPC)

	log.Print("gRPC сервер создан")

	errCh := make(chan error, 1)

	go func() {
		log.Print("gRPC сервер запускается " + lis.Addr().String() + "...")
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			log.Print("получен сигнал завершения, останавливаем gRPC сервер...")
		case err := <-errCh:
			log.Printf("gRPC сервер аварийно остановился: %s", err)
			return
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		done := make(chan struct{})
		go func() {
			grpcSrv.GracefulStop()
			close(done)
		}()

		select {
		case <-done:
			log.Print("[shutdown] gRPC сервер завершился корректно...")
		case <-shutdownCtx.Done():
			log.Print("[shutdown] таймаут graceful shutdown gRPC, вызываем Stop()")
			grpcSrv.Stop()
		}
	}()
}
