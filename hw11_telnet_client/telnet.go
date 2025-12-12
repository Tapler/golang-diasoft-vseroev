package main

import (
	"fmt"
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

type telnetClient struct {
	address string
	timeout time.Duration
	in      io.ReadCloser
	out     io.Writer
	conn    net.Conn
}

func NewTelnetClient(address string, timeout time.Duration, in io.ReadCloser, out io.Writer) TelnetClient {
	return &telnetClient{
		address: address,
		timeout: timeout,
		in:      in,
		out:     out,
	}
}

// Connect устанавливает TCP соединение с указанным адресом.
func (t *telnetClient) Connect() error {
	conn, err := net.DialTimeout("tcp", t.address, t.timeout)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}

// Close закрывает соединение.
func (t *telnetClient) Close() error {
	if t.conn != nil {
		return t.conn.Close()
	}
	return nil
}

// Send читает данные из in и отправляет их в сокет.
func (t *telnetClient) Send() error {
	if t.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := io.Copy(t.conn, t.in)
	return err
}

// Receive читает данные из сокета и записывает их в out.
func (t *telnetClient) Receive() error {
	if t.conn == nil {
		return fmt.Errorf("connection not established")
	}
	_, err := io.Copy(t.out, t.conn)
	return err
}
