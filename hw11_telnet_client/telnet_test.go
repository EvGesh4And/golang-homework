package main

import (
	"bytes"
	"errors"
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTelnetClient(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout, err := time.ParseDuration("10s")
			require.NoError(t, err)

			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)
			require.NoError(t, client.Connect())
			defer func() { require.NoError(t, client.Close()) }()

			in.WriteString("hello\n")
			err = client.Send()
			require.NoError(t, err)

			err = client.Receive()
			require.NoError(t, err)
			require.Equal(t, "world\n", out.String())
		}()

		go func() {
			defer wg.Done()

			conn, err := l.Accept()
			require.NoError(t, err)
			require.NotNil(t, conn)
			defer func() { require.NoError(t, conn.Close()) }()

			request := make([]byte, 1024)
			n, err := conn.Read(request)
			require.NoError(t, err)
			require.Equal(t, "hello\n", string(request)[:n])

			n, err = conn.Write([]byte("world\n"))
			require.NoError(t, err)
			require.NotEqual(t, 0, n)
		}()

		wg.Wait()
	})

	t.Run("SIGINT", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		errChan := make(chan error, 1)

		go func() {
			defer wg.Done()

			in := &bytes.Buffer{}
			out := &bytes.Buffer{}

			timeout := 10 * time.Second
			client := NewTelnetClient(l.Addr().String(), timeout, io.NopCloser(in), out)

			if err := client.Connect(); err != nil {
				errChan <- err
				return
			}

			time.Sleep(100 * time.Millisecond)
			err := client.Close()
			errChan <- err
		}()

		go func() {
			defer wg.Done()
			conn, err := l.Accept()
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()
		}()

		wg.Wait()

		select {
		case err := <-errChan:
			require.NoError(t, err)
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for client to close")
		}
	})

	t.Run("EOF (Ctrl+D)", func(t *testing.T) {
		l, err := net.Listen("tcp", "127.0.0.1:")
		require.NoError(t, err)
		defer func() { require.NoError(t, l.Close()) }()

		var wg sync.WaitGroup
		wg.Add(2)

		errChan := make(chan error, 1)

		// Клиентская сторона
		go func() {
			defer wg.Done()

			r, w := io.Pipe()
			_ = w.Close()

			out := &bytes.Buffer{}
			timeout := 10 * time.Second

			client := NewTelnetClient(l.Addr().String(), timeout, r, out)

			if err := client.Connect(); err != nil {
				errChan <- err
				return
			}

			err := client.Send()
			errChan <- err

			_ = client.Close()
		}()

		// Серверная сторона
		go func() {
			defer wg.Done()
			conn, err := l.Accept()
			require.NoError(t, err)
			defer func() { _ = conn.Close() }()
		}()

		wg.Wait()

		select {
		case err := <-errChan:
			if err != nil && !errors.Is(err, io.EOF) {
				t.Fatalf("unexpected error: %v", err)
			}
		case <-time.After(1 * time.Second):
			t.Fatal("timeout waiting for client to handle EOF")
		}
	})
}
