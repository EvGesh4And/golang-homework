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
		log.Fatal("Not all operands are specified")
	}
	host := args[0]
	port := args[1]
	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		log.Fatalf("Incorrect address: %v", err)
	}

	telClient := NewTelnetClient(addr.String(), *timeout, os.Stdin, os.Stdout)

	err = telClient.Connect()
	if err != nil {
		log.Fatalf("go-telnet: %v", err)
	}
	defer telClient.Close()

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)

	go func() {
		for {
			switch err := telClient.Send(); err {
			case nil:
			case io.EOF:
				log.Print("The input source is closed")
				cancel()
			default:
				log.Printf("Sending error: %v", err)
				cancel()
			}
		}
	}()
	go func() {
		for {
			switch err := telClient.Receive(); err {
			case nil:
			case io.EOF:
				log.Print("The server has closed the connection")
				cancel()
			default:
				log.Printf("Receiving error: %v", err)
				cancel()
			}
		}
	}()

	<-ctx.Done()
}
