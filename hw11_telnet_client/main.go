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
		fmt.Fprintln(os.Stderr, "Usage: go-telnet [--timeout duration] <host> <port>")
		pflag.PrintDefaults()
	}
	pflag.Parse()

	// Проверяем наличие двух позиционных аргументов: host и port
	args := pflag.Args()
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "not all operands are specified")
		return
	}
	host, port := args[0], args[1]

	addr, err := net.ResolveTCPAddr("tcp", host+":"+port)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}

	telClient := NewTelnetClient(addr.String(), *ptrTimeout, os.Stdin, os.Stdout)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	err = telClient.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	defer telClient.Close()
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", addr.String())

	go func() {
		defer cancel()
		if err := telClient.Send(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			select {
			case <-ctx.Done():
			default:
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}
	}()
	go func() {
		defer cancel()
		if err := telClient.Receive(); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			select {
			case <-ctx.Done():
			default:
				fmt.Fprintln(os.Stderr, err.Error())
			}
		}
	}()

	<-ctx.Done()

	fmt.Fprintf(os.Stderr, "...Connection to %s is closed\n", addr.String())
}
