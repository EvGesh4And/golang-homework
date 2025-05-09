package main

import (
	"context"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
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

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT)

	inputDone := make(chan struct{})
	errDone := make(chan struct{})
	go func() {
		for {
			switch err := telClient.Send(); err {
			case nil:
			case io.EOF:
				log.Print("go-telnet: input source is closed")
				close(inputDone)
				return
			default:
				log.Printf("go-telnet: sending error: %v", err)
				close(errDone)
				return
			}
		}
	}()
	go func() {
		for {
			switch err := telClient.Receive(); err {
			case nil:
			case io.EOF:
				log.Print("go-telnet: server has closed the connection")
				close(errDone)
				return
			default:
				log.Printf("go-telnet: receiving error: %v", err)
				close(errDone)
				return
			}
		}
	}()

	select {
	case <-ctx.Done():
		log.Print("go-telnet: interrupted by user (Ctrl+C)")
	case <-inputDone:
		log.Print("go-telnet: input closed by user (Ctrl+D)")
	case <-errDone:
		os.Exit(1)
	}
}
