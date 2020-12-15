package main

import (
	"context"
	"errors"
	"time"

	"github.com/extrame/go-socks5"
	"github.com/extrame/grpc-socks/pb"
	"github.com/sirupsen/logrus"
)

var waiterForSessions = make(map[string]map[pb.PayloadTypes]chan interface{})

type client struct {
	pb.Proxy_PumpServer
	clientCount int
}

func (c *client) ResolveIP(ctx context.Context, addr *pb.IPAddr) (*pb.IPAddr, error) {
	var payload pb.Payload
	payload.Type = pb.Payload_ResolveIP
	payload.Addr = addr
	payload.SessionId = ctx.Value(socks5.SessionID).(string)
	w := waitForSession(ctx.Value(socks5.SessionID).(string), pb.Payload_IPResolved)
	c.Send(&payload)
	select {
	case res := <-w:
		return res.(*pb.IPAddr), nil
	case <-time.After(10 * time.Second):
		return nil, errors.New("timeout")
	}
}

func (c *client) Connect(ctx context.Context, request *pb.Payload) error {
	request.SessionId = ctx.Value(socks5.SessionID).(string)
	request.Type = pb.Payload_Connect
	w := waitForSession(ctx.Value(socks5.SessionID).(string), pb.Payload_Connected)
	c.Send(request)
	select {
	case res := <-w:
		switch tr := res.(type) {
		case error:
			return tr
		}
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("timeout")
	}
}

func (c *client) Close(id string) error {
	var payload pb.Payload
	payload.SessionId = id
	payload.Type = pb.Payload_Close
	w := waitForSession(id, pb.Payload_Close)
	c.Send(&payload)
	select {
	case <-w:
		return nil
	case <-time.After(10 * time.Second):
		return errors.New("timeout")
	}
}

func waitForSession(id string, typ pb.PayloadTypes) chan interface{} {
	w, ok := waiterForSessions[id]
	if !ok {
		w = make(map[pb.PayloadTypes]chan interface{})
		waiterForSessions[id] = w
	}
	wtyp, ok := w[typ]
	if ok {
		wtyp <- errors.New("cancel by new waiter")
	}
	wtyp = make(chan interface{}, 0)
	waiterForSessions[id][typ] = wtyp
	logrus.Debugln(waiterForSessions)
	return wtyp
}
