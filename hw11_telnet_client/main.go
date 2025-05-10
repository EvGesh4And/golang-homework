package main

import (
	"context"
	"errors"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	// Place your code here,
	// P.S. Do not rush to throw context down, think think if it is useful with blocking operation?
	timeout := pflag.Duration("timeout", 10*time.Second, "timeout for connecting to the server")
	pflag.Usage = func() {
		log.Print("Usage: go-telnet [--timeout duration] <host> <port>")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	// Проверяем наличие двух позиционных аргументов: host и port
	args := pflag.Args()
	if len(args) < 2 {
		log.Fatal("go-telnet: not all operands are specified")
	}
	host := args[0]
	port := args[1]
	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("go-telnet: incorrect address: %v", err)
	}

	telClient := NewTelnetClient(addr.String(), *timeout, os.Stdin, os.Stdout)

	err = telClient.Connect()
	if err != nil {
		log.Fatalf("go-telnet: connection error: %v", err)
	}
	log.Printf("go-telnet: connected to %s", addr.String())
	defer telClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for {
			err := telClient.Send()
			switch {
			case err == nil:
				// всё ок
			case errors.Is(err, io.EOF):
				log.Print("go-telnet: input source is closed")
				cancel()
				return
			default:
				log.Printf("go-telnet: sending error: %v", err)
				cancel()
				return
			}
		}
	}()
	go func() {
		defer wg.Done()
		for {
			err := telClient.Receive()
			switch {
			case err == nil:
				// всё ок
			case errors.Is(err, io.EOF):
				log.Print("go-telnet: server has closed the connection")
				cancel()
				return
			default:
				log.Printf("go-telnet: receiving error: %v", err)
				cancel()
				return
			}
		}
	}()
	wg.Wait()
	// Ожидаем отмену
	<-ctx.Done()
}
