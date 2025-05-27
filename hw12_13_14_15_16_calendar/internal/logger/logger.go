// Package logger предоставляет средства для логирования,
// включая уровни логов, форматирование и вывод в разные источники.
package logger

import (
	"fmt"
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
}

func New(level string) *Logger {
	lvl, ok := levelMap[level]
	if !ok {
		// Дефолтный уровень
		lvl = LevelInfo
	}
	return &Logger{level: lvl}
}

func (l Logger) logPrint(level int, levelName string, msg string) {
	if level > l.level {
		return
	}
	fmt.Printf("[%s] %s %s\n", levelName, time.Now().Format("2006-01-02 15:04:05.000"), msg)
}

func (l Logger) Error(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelError, "ERROR", msg)
}

func (l Logger) Warn(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelWarn, "WARN", msg)
}

func (l Logger) Info(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelInfo, "INFO", msg)
}

func (l Logger) Debug(form string, args ...any) {
	msg := fmt.Sprintf(form, args...)
	l.logPrint(LevelDebug, "DEBUG", msg)
}
