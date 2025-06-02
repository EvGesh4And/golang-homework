// Package logger предоставляет средства для логирования,
// включая уровни логов, форматирование и вывод в разные источники.
package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

var levelMap = map[string]slog.Level{
	"error": slog.LevelError,
	"warn":  slog.LevelWarn,
	"info":  slog.LevelInfo,
	"debug": slog.LevelDebug,
}

func New(level string, out io.Writer) *slog.Logger {
	if lvl, ok := levelMap[level]; ok {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: lvl,
		})
		logger := slog.New(handler)
		log.Print("уровень логгирования: ", level)
		return logger
	} else {
		handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
		logger := slog.New(handler)
		log.Print("уровень логгирования: debug (по умолчанию)")
		return logger
	}
}
