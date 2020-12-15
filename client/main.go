package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"time"

	// socks5 "github.com/armon/go-socks5"
	// "github.com/coocood/freecache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/sirupsen/logrus"

	// "google.golang.org/grpc/resolver"

	"github.com/extrame/grpc-socks/lib"
	// "github.com/extrame/grpc-socks/log"
	"github.com/extrame/grpc-socks/pb"
)

var (
	debug    = false
	compress = false

	addr        = "127.0.0.1"
	port        = 50051
	callOptions = make([]grpc.CallOption, 0)

	proxyClient pb.ProxyClient
	tolerant    uint
	period      uint

	connections = make(map[string]net.Conn)
)

func init() {
	flag.BoolVar(&debug, "d", debug, "debug mode")
	flag.StringVar(&addr, "l", addr, "local addr")
	flag.IntVar(&port, "p", 50051, "local port")
	flag.BoolVar(&compress, "cp", compress, "enable snappy compress")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
	// if debug {
	// 	logrus.SetDebugMode()
	// }

	if compress {
		encoding.RegisterCompressor(lib.Snappy())
		callOptions = append(callOptions, grpc.UseCompressor("snappy"))
	}

}

func main() {
	// resolver.Register(&etcdResolver{})

	//作为连接前端，连接服务器，等待请求
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", addr, port), grpc.WithTransportCredentials(lib.ClientTLS(addr)))
	if err != nil {
		panic(err)
	}
	proxyClient = pb.NewProxyClient(conn)

	//等待接入
	steam, err := proxyClient.Pump(context.Background(), callOptions...)

	if err == nil {
		for {
			s, err := steam.Recv()
			if err != nil {
				logrus.Fatal(err)
			}
			switch s.Type {
			case pb.Payload_ResolveIP:
				ipAddr, err := net.ResolveIPAddr("ip", s.Addr.Address)
				var result pb.Payload
				if err == nil {
					result.SessionId = s.SessionId
					result.Type = pb.Payload_IPResolved
					result.Addr = &pb.IPAddr{
						Data: ipAddr.IP,
						Zone: ipAddr.Zone,
					}
				}
				steam.Send(&result)
			case pb.Payload_Connect:
				logrus.Infoln("connect...", s)
				if s.Addr != nil {
					conn, err := net.DialTimeout("tcp", s.Addr.Address, time.Second*15)
					if err != nil {
						logrus.Errorf("tcp dial %q err: %s", s.Addr.Address, err)
					}

					conn.(*net.TCPConn).SetKeepAlive(true)
					connections[s.SessionId] = conn
					go (&session{id: s.SessionId, conn: conn}).watch(steam)
					logrus.Debugf("tcp conn %q<-->%q<-->%q", "remote", addr, conn.RemoteAddr())
					steam.Send(&pb.Payload{
						SessionId: s.SessionId,
						Type:      pb.Payload_Connected,
					})
				} else {
					steam.Send(&pb.Payload{
						Error: "receive addr nil",
						Type:  pb.Payload_ConnectErr,
					})
				}

			case pb.Payload_Close:
				conn, ok := connections[s.SessionId]
				if ok {
					conn.Close()
				}
				steam.Send(&pb.Payload{
					SessionId: s.SessionId,
					Type:      pb.Payload_Closed,
				})
			case pb.Payload_DATA:
				conn, ok := connections[s.SessionId]
				if ok {
					conn.Write(s.Data)
				}
			}
		}
	} else {
		logrus.Errorln(err)
	}

}
