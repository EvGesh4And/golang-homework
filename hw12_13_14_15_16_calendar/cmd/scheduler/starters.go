package main

import (
	"context"
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/scheduler"
	sqlstorage "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/storage/sql"
)

func setupStorage(ctx context.Context, cfg Config, lg *slog.Logger) (scheduler.Storage, io.Closer, error) {
	log.Print("initializing connection to PostgreSQL...")

	sqlStorage := sqlstorage.New(lg, cfg.Storage.DSN)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := sqlStorage.Connect(ctx); err != nil {
		log.Printf("error connecting to PostgreSQL: %v", err)
		return nil, nil, err
	}

	log.Print("sql storage initialized and connected successfully")
	return sqlStorage, sqlStorage, nil
}
