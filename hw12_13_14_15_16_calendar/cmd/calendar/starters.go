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
			log.Printf("failed to open log file %s: %s", filePath, err.Error())
			return nil, nil, err
		}
		globalLogger = logger.New(cfg.Logger.Level, logFile)
	default:
		log.Printf("unknown logger mode: %s, using console", cfg.Logger.Mod)
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
		log.Print("using in-memory storage")
		return memorystorage.New(logStorageMem), nil, nil

	case "sql":
		log.Print("initializing connection to PostgreSQL...")

		sqlStorage := sqlstorage.New(logStorageSQL, cfg.Storage.DSN)
		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if err := sqlStorage.Connect(ctx); err != nil {
			log.Printf("failed to connect to PostgreSQL: %v", err)
			return nil, nil, err
		}

		log.Print("running migrations...")
		if err := sqlStorage.Migrate(cfg.Storage.Migration); err != nil {
			log.Print(err)
			return nil, nil, err
		}

		log.Print("SQL storage initialized and connected successfully")
		return sqlStorage, sqlStorage, nil

	default:
		log.Printf("unknown storage type: %v", cfg.Storage.Mod)
		return nil, nil, fmt.Errorf("unknown storage type: %v", cfg.Storage.Mod)
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
	log.Print("HTTP server created")

	addr := fmt.Sprintf("%s:%d", cfg.HTTP.Host, cfg.HTTP.Port)
	errCh := make(chan error, 1)

	go func() {
		log.Print("HTTP server starting " + addr + "...")
		if err := serverHTTP.Start(); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			log.Print("shutdown signal received, stopping HTTP server...")
		case err := <-errCh:
			log.Printf("HTTP server stopped unexpectedly: %s", err)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		if err := serverHTTP.Stop(shutdownCtx); err != nil {
			log.Printf("[shutdown] HTTP server shutdown error: %s", err)
		} else {
			log.Print("[shutdown] HTTP server shut down gracefully...")
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
		log.Printf("gRPC failed to listen on %v: %v", addr, err)
		os.Exit(1)
	}

	serverGRPC := grpcserver.NewServerGRPC(logGRPC, lis, calendar)
	grpcSrv := grpc.NewServer()
	pb.RegisterCalendarServer(grpcSrv, serverGRPC)

	log.Print("gRPC server created")

	errCh := make(chan error, 1)

	go func() {
		log.Print("gRPC server starting " + lis.Addr().String() + "...")
		if err := grpcSrv.Serve(lis); err != nil {
			errCh <- err
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		select {
		case <-ctx.Done():
			log.Print("shutdown signal received, stopping gRPC server...")
		case err := <-errCh:
			log.Printf("gRPC server stopped unexpectedly: %s", err)
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
			log.Print("[shutdown] gRPC server shut down gracefully...")
		case <-shutdownCtx.Done():
			log.Print("[shutdown] graceful shutdown timeout for gRPC, calling Stop()")
			grpcSrv.Stop()
		}
	}()
}
