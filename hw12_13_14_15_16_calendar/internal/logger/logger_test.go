package logger

import (
	"bytes"
	"io"
	"os"
	"strings"
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
		log.Debug("test", "проверка валидности события с ID: %d", 10)
	})

	require.Contains(t, output, "DEBUG", "должен содержать уровень DEBUG")
	require.Contains(t, output, "проверка валидности события с ID: 10", "должен содержать сообщение")

	output = captureOutput(func(w io.Writer) {
		log := New("info", w)
		log.Info("test", "добавлено событие с ID: %d", 10)
		log.Warn("test", "потеряно соединение, попытка его восстановить")
	})

	slice := strings.Split(output, "\n")

	require.Len(t, slice, 3, "число логов не соотвествует требованию")

	require.Contains(t, slice[0], "INFO", "должен содержать уровень INFO")
	require.Contains(t, slice[0], "добавлено событие с ID: 10", "должен содержать сообщение")

	require.Contains(t, slice[1], "WARN", "должен содержать уровень WARN")
	require.Contains(t, slice[1], "потеряно соединение, попытка его восстановить", "должен содержать сообщение")

	output = captureOutput(func(w io.Writer) {
		log := New("warn", w)
		log.Info("test", "добавлено событие с ID: %d", 10)
		log.Warn("test", "потеряно соединение, попытка его восстановить")
	})

	slice = strings.Split(output, "\n")

	require.Len(t, slice, 2, "число логов не соотвествует требованию")

	require.Contains(t, slice[0], "WARN", "должен содержать уровень WARN")
	require.Contains(t, slice[0], "потеряно соединение, попытка его восстановить", "должен содержать сообщение")

	output = captureOutput(func(w io.Writer) {
		log := New("error", w)
		log.Info("test", "добавлено событие с ID: %d", 10)
		log.Error("test", "связь с БД полностью потеряно")
		log.Warn("test", "потеряно соединение, попытка его восстановить")
		log.Debug("test", "проверка валидности события с ID: %d", 10)
	})

	slice = strings.Split(output, "\n")

	require.Len(t, slice, 2, "число логов не соотвествует требованию")

	require.Contains(t, slice[0], "ERROR", "должен содержать уровень WARN")
	require.Contains(t, slice[0], "связь с БД полностью потеряно", "должен содержать сообщение")
}
