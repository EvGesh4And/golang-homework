// Package logger предоставляет средства для логирования,
// включая уровни логов, форматирование и вывод в разные источники.
package logger

import (
	"fmt"
	"io"
	"log"
)

// Константы уровней.
const (
	LevelError = iota
	LevelWarn
	LevelInfo
	LevelDebug
)

var levelMap = map[string]int{
	"error": LevelError,
	"warn":  LevelWarn,
	"info":  LevelInfo,
	"debug": LevelDebug,
}

type Logger struct {
	level int
	std   *log.Logger
}

func New(level string, out io.Writer) *Logger {
	lvl, ok := levelMap[level]
	if !ok {
		lvl = LevelInfo
	}
	return &Logger{
		level: lvl,
		std:   log.New(out, "", log.LstdFlags|log.Lmicroseconds),
	}
}

func (l *Logger) logPrint(level int, levelName string, msg string) {
	if level > l.level {
		return
	}
	l.std.Printf("[%s] %s\n", levelName, msg)
}

func (l *Logger) Error(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelError, "ERROR", msg)
}

func (l *Logger) Warn(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelWarn, "WARN", msg)
}

func (l *Logger) Info(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelInfo, "INFO", msg)
}

func (l *Logger) Debug(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelDebug, "DEBUG", msg)
}
