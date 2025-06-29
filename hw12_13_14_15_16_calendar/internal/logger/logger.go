// Package logger предоставляет средства для логирования,
// включая уровни логов, форматирование и вывод в разные источники.
package logger

import (
	"io"
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

// New creates a logger with the provided level and writer.
func New(level string, out io.Writer) *slog.Logger {
	var levLog slog.Level
	if lvl, ok := levelMap[level]; ok {
		levLog = lvl
	} else {
		levLog = slog.LevelDebug
	}

	// color := isStdout(out)

	handler := tint.NewHandler(out, &tint.Options{
		Level:      levLog,
		TimeFormat: time.Kitchen,
		// NoColor:    !color,
		NoColor: true,
	})

	handler = NewHandlerMiddleware(handler)

	return slog.New(handler)
}

// func isStdout(out io.Writer) bool {
// 	f, ok := out.(*os.File)
// 	if !ok {
// 		return false
// 	}
// 	return f.Fd() == os.Stdout.Fd()
// }
