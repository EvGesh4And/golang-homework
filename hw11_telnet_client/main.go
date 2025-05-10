package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/pflag"
)

func main() {
	ptrTimeout := pflag.Duration("timeout", 10*time.Second, "timeout for connecting to the server")
	pflag.Usage = func() {
		fmt.Fprint(os.Stderr, "Usage: go-telnet [--timeout duration] <host> <port>")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	// Проверяем наличие двух позиционных аргументов: host и port
	args := pflag.Args()
	if len(args) < 2 {
		fmt.Fprint(os.Stderr, "not all operands are specified")
	}
	host, port := args[0], args[1]

	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}

	telClient := NewTelnetClient(addr.String(), *ptrTimeout, os.Stdin, os.Stdout)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)

	err = telClient.Connect()
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	defer telClient.Close()
	fmt.Fprintf(os.Stderr, "...Connected to %s", addr.String())

	go func() {
		defer cancel()
		if err := telClient.Send(); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
	}()
	go func() {
		defer cancel()
		if err := telClient.Receive(); err != nil {
			fmt.Fprint(os.Stderr, err.Error())
			return
		}
	}()
	<-ctx.Done()
}
