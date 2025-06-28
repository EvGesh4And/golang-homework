package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	logsetup "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger/setup"
	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
)

type ChildLoggers struct {
	scheduler  *slog.Logger
	storageSQL *slog.Logger
}

func setupLogger(cfg Config) (*ChildLoggers, io.Closer, error) {
	globalLogger, closer, err := logsetup.New(logsetup.Config{
		Mod:   cfg.Logger.Mod,
		Path:  cfg.Logger.Path,
		Level: cfg.Logger.Level,
	})
	if err != nil {
		return nil, nil, err
	}

	childLoggers := &ChildLoggers{
		scheduler:  globalLogger.With("component", "scheduler"),
		storageSQL: globalLogger.With("component", "storage", "type", "sql"),
	}

	return childLoggers, closer, nil
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
