package main

import (
	"net"

	"github.com/extrame/grpc-socks/lib"
	"github.com/extrame/grpc-socks/pb"
	"github.com/sirupsen/logrus"
)

const leakyBufSize = 4108 // data.len(2) + hmacsha1(10) + data(4096)

var leakyBuf = lib.NewLeakyBuf(2048, leakyBufSize)

type session struct {
	id   string
	conn net.Conn
}

func (s *session) watch(client pb.Proxy_PumpClient) {
	buff := leakyBuf.Get()
	defer leakyBuf.Put(buff)
	for {
		n, err := s.conn.Read(buff)
		if err != nil {
			break
		}

		if n > 0 {
			var result pb.Payload
			result.SessionId = s.id
			result.Data = buff[:n]
			err = client.Send(&result)
			if err != nil {
				logrus.Errorf("stream send err: %s", err)
				break
			}
		}
	}
}
