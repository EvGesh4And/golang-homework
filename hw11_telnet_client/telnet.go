package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"time"
)

type TelnetClient interface {
	Connect() error
	io.Closer
	Send() error
	Receive() error
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &MyTelnetClinet{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

type MyTelnetClinet struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
	ctx     context.Context
}

func (mtc *MyTelnetClinet) Connect() error {
	ctx, _ := context.WithTimeout(context.Background(), mtc.timeout)
	// defer cancel()

	dial := net.Dialer{}
	conn, err := dial.DialContext(ctx, "tcp", mtc.address)
	if err != nil {
		return err
	}
	// defer conn.Close()
	log.Printf("Connected to %s", mtc.address)
	mtc.conn = conn
	mtc.ctx = ctx
	return nil
}

func (mtc *MyTelnetClinet) Close() error {
	if mtc.conn != nil {
		err := mtc.conn.Close()
		if err != nil {
			return err
		}
		err = mtc.in.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (mtc *MyTelnetClinet) Send() error {
	scanner := bufio.NewScanner(mtc.in)

	for scanner.Scan() {
		_, err := mtc.conn.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}

func (mtc *MyTelnetClinet) Receive() error {
	scanner := bufio.NewScanner(mtc.conn)
	for scanner.Scan() {
		_, err := mtc.out.Write([]byte(scanner.Text() + "\n"))
		if err != nil {
			return err
		}
	}
	return nil
}
