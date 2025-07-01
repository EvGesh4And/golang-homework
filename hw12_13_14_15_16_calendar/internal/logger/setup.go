package logger

import (
	"io"
	"log"
	"log/slog"
	"os"
)

// Config describes logger initialization parameters.
type Config struct {
	Mod   string `toml:"mod" env:"MOD"`
	Path  string `toml:"path" env:"PATH"`
	JSON  bool   `toml:"json" env:"JSON"`
	Level string `toml:"level" env:"LEVEL"`
}

// New initializes global slog.Logger according to configuration.
// It returns created logger and optional io.Closer that should be closed
// when logger output is a file.
func NewLogger(cfg Config) (*slog.Logger, io.Closer, error) {
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

	l := New(cfg.Level, out, cfg.JSON)
	var closer io.Closer
	if c, ok := out.(io.Closer); ok && c != os.Stdout {
		closer = c
	}
	return l, closer, nil
}
