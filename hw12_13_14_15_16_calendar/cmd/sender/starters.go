package main

import (
	"io"
	"log"
	"log/slog"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
	logsetup "github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger/setup"
)

type ChildLoggers struct {
	sender *slog.Logger
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
		sender: globalLogger.With("component", "sender"),
	}

	return childLoggers, closer, nil
}
