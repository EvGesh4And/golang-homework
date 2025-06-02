package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func captureOutput(f func(w io.Writer)) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f(w)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestLogger_Info(t *testing.T) {
	output := captureOutput(func(w io.Writer) {
		log := New("debug", w)
		log.Debug("проверка валидности события", "event_id", 10)
	})

	require.Contains(t, output, "DEBUG", "должен содержать уровень DEBUG")
	require.Contains(t, output, "проверка валидности события", "должен содержать сообщение")

	output = captureOutput(func(w io.Writer) {
		log := New("info", w)
		log.Info("добавлено событие", "event_id", 10)
		log.Warn("потеряно соединение, попытка его восстановить")
	})

	require.Contains(t, output, "INFO", "должен содержать уровень INFO")
	require.Contains(t, output, "добавлено событие", "должен содержать сообщение")

	require.Contains(t, output, "WARN", "должен содержать уровень WARN")
	require.Contains(t, output, "потеряно соединение, попытка его восстановить", "должен содержать сообщение")

	output = captureOutput(func(w io.Writer) {
		log := New("warn", w)
		log.Info("добавлено событие", "event_id", 10)
		log.Warn("потеряно соединение, попытка его восстановить")
	})

	require.Contains(t, output, "WARN", "должен содержать уровень WARN")
	require.Contains(t, output, "потеряно соединение, попытка его восстановить", "должен содержать сообщение")

	output = captureOutput(func(w io.Writer) {
		log := New("error", w)
		log.Info("добавлено событие", "event_id", 10)
		log.Error("связь с БД полностью потеряно")
		log.Warn("потеряно соединение, попытка его восстановить")
		log.Debug("проверка валидности события", "event_id", 10)
	})
	fmt.Println(output)
	require.Contains(t, output, "ERROR", "должен содержать уровень ERROR")
	require.Contains(t, output, "связь с БД полностью потеряно", "должен содержать сообщение")
}
