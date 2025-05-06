package main

import (
	"log"
	"net"
	"os"
	"sync"
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
	log.Print(addr.String())
	err = telClient.Connect()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		telClient.Send()
	}()
	go func() {
		defer wg.Done()
		go telClient.Receive()
	}()
	wg.Wait()
}
