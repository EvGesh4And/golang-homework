// Package logger предоставляет средства для логирования,
// включая уровни логов, форматирование и вывод в разные источники.
package logger

import (
	"io"
	"log"
	"log/slog"
	"time"

	"github.com/lmittmann/tint"
)

var levelMap = map[string]slog.Level{
	"error": slog.LevelError,
	"warn":  slog.LevelWarn,
	"info":  slog.LevelInfo,
	"debug": slog.LevelDebug,
}

func New(level string, out io.Writer) *slog.Logger {
	var levLog slog.Level

	if lvl, ok := levelMap[level]; ok {
		levLog = lvl
		log.Print("уровень логгирования: ", lvl)
	} else {
		levLog = slog.LevelDebug
		log.Print("уровень логгирования: debug (по умолчанию)")
	}

	handler := tint.NewHandler(out, &tint.Options{
		Level:      levLog,
		TimeFormat: time.Kitchen,
	})

	handler = NewHandlerMiddleware(handler)

	logger := slog.New(handler)

	return logger
}
