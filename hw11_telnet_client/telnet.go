package main

import (
	"io"
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
	mtc.conn = conn
	return nil
}

func (mtc *MyTelnetClinet) Send() error {
	_, err := io.Copy(mtc.conn, mtc.in)
	if err != nil {
		return err
	}
	return nil
}

func (mtc *MyTelnetClinet) Receive() error {
	_, err := io.Copy(mtc.out, mtc.conn)
	if err != nil {
		return err
	}
	return nil
}

func (mtc *MyTelnetClinet) Close() error {
	return mtc.conn.Close()
}
