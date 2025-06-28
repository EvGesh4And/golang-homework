package setup

import (
	"io"
	"log"
	"log/slog"
	"os"

	"github.com/EvGesh4And/golang-homework/hw12_13_14_15_16_calendar/internal/logger"
)

// Config describes logger initialization parameters.
type Config struct {
	Mod   string
	Path  string
	Level string
}

// New initializes global slog.Logger according to configuration.
// It returns created logger and optional io.Closer that should be closed
// when logger output is a file.
func New(cfg Config) (*slog.Logger, io.Closer, error) {
	var out io.WriteCloser
	switch cfg.Mod {
	case "console", "":
		out = os.Stdout
	case "file":
		filePath := cfg.Path
		if filePath == "" {
			filePath = "calendar.log"
		}
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o666)
		if err != nil {
			log.Printf("error opening log file %s: %s", filePath, err)
			return nil, nil, err
		}
		out = f
	default:
		log.Printf("unknown logger mode: %s, using console", cfg.Mod)
		out = os.Stdout
	}

	l := logger.New(cfg.Level, out)
	var closer io.Closer
	if c, ok := out.(io.Closer); ok && c != os.Stdout {
		closer = c
	}
	return l, closer, nil
}
