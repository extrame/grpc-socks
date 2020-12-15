package main

import (
	"fmt"
	"net"
	"time"

	"github.com/extrame/grpc-socks/pb"
)

type session struct {
	id     string
	addr   *net.TCPAddr
	client *client
	bytes  []byte
}

func (c *session) Read(b []byte) (n int, err error) {
	var length = len(c.bytes)
	if len(c.bytes) >= len(b) {
		length = len(b)
	}
	copy(b, c.bytes[:length])
	c.bytes = c.bytes[length:]
	return length, nil
}

func (c *session) watch() {
	for {
		w := waitForSession(c.id, pb.Payload_DATA)
		res := <-w
		switch tr := res.(type) {
		case []byte:
			c.bytes = append(c.bytes, tr...)
		default:
			fmt.Print(".")
		}
	}

}

func (c *session) Write(b []byte) (n int, err error) {
	p := &pb.Payload{
		SessionId: c.id,
		Data:      b,
	}

	return len(b), c.client.Send(p)
}

func (c *session) Close() error {
	c.client.Close(c.id)
	delete(waiterForSessions, c.id)
	return nil
}

func (c *session) LocalAddr() net.Addr {
	return c.addr
}

func (c *session) RemoteAddr() net.Addr {
	return nil
}

// TODO impl
func (c *session) SetDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *session) SetReadDeadline(t time.Time) error {
	return nil
}

// TODO impl
func (c *session) SetWriteDeadline(t time.Time) error {
	return nil
}
