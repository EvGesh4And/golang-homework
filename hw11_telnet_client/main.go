package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
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

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT)

	err = telClient.Connect()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		return
	}
	defer telClient.Close()
	fmt.Fprintf(os.Stderr, "...Connected to %s\n", addr.String())
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go dd(ctx, wg, telClient.Send, cancel)
	go dd(ctx, wg, telClient.Receive, cancel)

	<-ctx.Done()
	fmt.Fprintln(os.Stderr, "...Interrupt received, closing connection")
	wg.Wait()
	fmt.Fprintf(os.Stderr, "...Connection to %s is closed\n", addr.String())
}

func dd(ctx context.Context, wg *sync.WaitGroup, f func() error, cancel context.CancelFunc) {
	defer wg.Done()
	defer cancel()
	if err := f(); err != nil {
		select {
		case <-ctx.Done():
		default:
			fmt.Fprintln(os.Stderr, err.Error())
		}
		return
	}
}

// go func() {
// 	defer wg.Done()
// 	defer cancel()
// 	if err := telClient.Send(); err != nil {
// 		select {
// 		case <-ctx.Done():
// 		default:
// 			fmt.Fprintln(os.Stderr, err.Error())
// 		}
// 		return
// 	}
// }()
