// Package logger предоставляет средства для логирования,
// включая уровни логов, форматирование и вывод в разные источники.
package logger

import (
	"fmt"
	"io"
	"log"
	"time"
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
		std:   log.New(out, "", 0),
	}
}

func (l *Logger) logPrint(level int, levelName string, module string, msg string) {
	if level > l.level {
		return
	}
	// Дата и время.
	ts := time.Now().Format("02/Jan/2006:15:04:05 -0700")
	l.std.Printf("%-16s [%s] %s: %s\n", levelName, ts, module, msg)
}

func (l *Logger) Error(module string, form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelError, "ERROR", module, msg)
}

func (l *Logger) Warn(module string, form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelWarn, "WARN", module, msg)
}

func (l *Logger) Info(module string, form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelInfo, "INFO", module, msg)
}

func (l *Logger) Debug(module string, form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelDebug, "DEBUG", module, msg)
}

func (l *Logger) Printf(form string, args ...any) {
	l.std.Printf(form, args...)
}
