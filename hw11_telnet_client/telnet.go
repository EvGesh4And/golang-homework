package main

import (
	"bufio"
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
}

func (mtc *MyTelnetClinet) Connect() error {

	conn, err := net.DialTimeout("tcp", mtc.address, mtc.timeout)
	if err != nil {
		return err
	}
	log.Printf("Connected to %s", mtc.address)
	mtc.conn = conn
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
	log.Print("Stop Send")
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
	log.Print("Stop Receive")
	return nil
}
