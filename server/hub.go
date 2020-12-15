package main

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"github.com/extrame/grpc-socks/lib"
	"github.com/extrame/grpc-socks/pb"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/peer"
)

type hub struct {
	serverToken []byte
	connected   map[string]*client
}

func (h *hub) Echo(ctx context.Context, req *pb.Payload) (*pb.Payload, error) {
	return &pb.Payload{Data: h.serverToken}, nil
}

func (h *hub) Pump(stream pb.Proxy_PumpServer) error {
	//客户端登记自身能力，等待使用
	peer, _ := peer.FromContext(stream.Context())
	logrus.Info("get client from ", peer.Addr)
	h.connected[peer.Addr.String()] = &client{stream, 0}
	var err error
	for {
		var pay *pb.Payload
		pay, err = stream.Recv()
		if err == nil {
			logrus.Debugln("get message", pay)
			switch pay.Type {
			case pb.Payload_IPResolved:
				waiterForSessions[pay.SessionId][pb.Payload_IPResolved] <- pay.Addr
			case pb.Payload_Closed, pb.Payload_Connected:
				waiterForSessions[pay.SessionId][pb.Payload_Connected] <- true
			case pb.Payload_ConnectErr:
				logrus.WithField("type", pb.Payload_ConnectErr.String).WithField("error", pay.Error).Errorln("connect fail")
				waiterForSessions[pay.SessionId][pb.Payload_Connected] <- errors.New(pay.Error)
			case pb.Payload_DATA:
				waiterForSessions[pay.SessionId][pb.Payload_DATA] <- pay.Data
			}
		} else {
			break
		}
	}
	delete(h.connected, peer.Addr.String())
	logrus.Errorln("client disconnected by ", err)
	return nil
}

func (h *hub) PipelineUDP(stream pb.Proxy_PipelineUDPServer) error {
	frame := &pb.Payload{}

	err := stream.RecvMsg(frame)
	if err != nil {
		logrus.Errorf("udp first frame err: %s", err)
		return err
	}

	addr := string(frame.Data)

	logrus.Debugf("recv udp addr: %s", addr)

	conn, err := net.Dial("udp", addr)
	if err != nil {
		logrus.Errorf("udp dial %s err: %s", addr, err)
		return err
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Second * 600))

	ctx := stream.Context()

	go func() {
		buff := make([]byte, lib.UDPMaxSize)

		for {
			n, err := conn.Read(buff)
			if n > 0 {
				frame.Data = buff[:n]
				err = stream.Send(frame)
				if err != nil {
					logrus.Errorf("stream send err: %s", err)
					break
				}
			}

			if err != nil {
				break
			}
		}
	}()

	for {
		p, err := stream.Recv()

		if err == io.EOF {
			return nil
		}

		if err != nil {
			if ctx.Err() == context.Canceled {
				break
			}
			logrus.Errorf("stream recv err: %s", err)
			return err
		}

		_, err = conn.Write(p.Data)
		if err != nil {
			logrus.Errorf("udp conn write err: %s", err)
			return err
		}
	}

	return nil
}
