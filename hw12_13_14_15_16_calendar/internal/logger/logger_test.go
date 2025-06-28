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
		log.Debug("checking event validity", "event_id", 10)
	})

	require.Contains(t, output, "checking event validity", "should contain message")

	output = captureOutput(func(w io.Writer) {
		log := New("info", w)
		log.Info("event added", "event_id", 10)
		log.Warn("connection lost, attempting to restore")
	})

	require.Contains(t, output, "event added", "should contain message")

	require.Contains(t, output, "connection lost, attempting to restore", "should contain message")

	output = captureOutput(func(w io.Writer) {
		log := New("warn", w)
		log.Info("event added", "event_id", 10)
		log.Warn("connection lost, attempting to restore")
	})

	require.Contains(t, output, "connection lost, attempting to restore", "should contain message")

	output = captureOutput(func(w io.Writer) {
		log := New("error", w)
		log.Info("event added", "event_id", 10)
		log.Error("database connection completely lost")
		log.Warn("connection lost, attempting to restore")
		log.Debug("checking event validity", "event_id", 10)
	})
	fmt.Println(output)
	require.Contains(t, output, "database connection completely lost", "should contain message")
}
