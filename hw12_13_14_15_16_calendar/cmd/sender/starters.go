package main

import (
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
)

type ChildLoggers struct {
	scheduler *slog.Logger
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
		scheduler: globalLogger.With("component", "sender"),
	}

	return childLoggers, logFile, nil
}
