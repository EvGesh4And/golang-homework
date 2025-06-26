package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
)

type ChildLoggers struct {
	scheduler  *slog.Logger
	storageSQL *slog.Logger
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
			log.Printf("error opening log file %s: %s", filePath, err.Error())
			return nil, nil, err
		}
		globalLogger = logger.New(cfg.Logger.Level, logFile)
	default:
		log.Printf("unknown logger mode: %s, using console", cfg.Logger.Mod)
		globalLogger = logger.New(cfg.Logger.Level, os.Stdout)
	}

	childLoggers := &ChildLoggers{
		scheduler:  globalLogger.With("component", "scheduler"),
		storageSQL: globalLogger.With("component", "storage", "type", "sql"),
	}

	return childLoggers, logFile, nil
}

func setupStorage(ctx context.Context, cfg Config, childLoggers *ChildLoggers) (scheduler.Storage, io.Closer, error) {
	logStorageSQL := childLoggers.storageSQL

	log.Print("initializing connection to PostgreSQL...")

	sqlStorage := sqlstorage.New(logStorageSQL, cfg.Storage.DSN)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlStorage.Connect(ctx); err != nil {
		log.Printf("error connecting to PostgreSQL: %v", err)
		return nil, nil, err
	}

	log.Print("sql storage initialized and connected successfully")
	return sqlStorage, sqlStorage, nil
}
