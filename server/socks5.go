package main

import (
	"context"
	"net"

	socks5 "github.com/extrame/go-socks5"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/extrame/grpc-socks/pb"
)

var callOptions = make([]grpc.CallOption, 0)
var preferedRemote string

//存储访问客户和直接转发远端的对应关系
var onlineClients = make(map[string]string)

func DialFunc(ctx context.Context, network, addr string) (net.Conn, error) {
	// log.Debugf("%q<-%s->%q", ctx.Value(nameCtxKey), network, addr)

	tcpAddr, err := net.ResolveTCPAddr(network, addr)
	if err != nil {
		return nil, err
	}

	//不需要进行转发的流量
	if isLocal(tcpAddr) {
		return net.DialTCP(network, nil, tcpAddr)
	}

	//从client里面选择合适的客户端分配给该请求
	var grpcClient = selectClient(ctx.Value(socks5.ClientID).(string))

	// ctx = metadata.AppendToOutgoingContext(ctx, "url", ctx.Value(nameCtxKey).(string))

	logrus.WithField("addr", tcpAddr.String()).Infoln("connect...")
	err = grpcClient.Connect(ctx, &pb.Payload{
		Addr: &pb.IPAddr{Address: tcpAddr.String()},
	})
	logrus.Debugln("connect", err)
	if err != nil {
		return nil, err
	}

	session := session{id: ctx.Value(socks5.SessionID).(string), addr: tcpAddr, client: grpcClient}
	go session.watch()
	return &session, nil
}

func selectClient(id string) (selected *client) {
	if c, ok := onlineClients[id]; ok {
		if remote, ok := grpcHub.connected[c]; ok {
			selected = remote
			goto returnClient
		}
	}
	//use prefered
	if preferedRemote != "" {
		if remote, ok := grpcHub.connected[preferedRemote]; ok {
			selected = remote
			goto returnClient
		} else {
			preferedRemote = ""
		}
	}
	//select from connected
	for k, v := range grpcHub.connected {
		if v.clientCount > maxPerRemote {
			logrus.Infoln("some remote exceed")
		} else {
			preferedRemote = k
			selected = v
			goto returnClient
		}
	}
	return nil
returnClient:
	selected.clientCount++
	return selected
}

// type client struct {
// 	addr   net.Addr
// 	stream pb.Proxy_PumpClient
// }

func isLocal(addr *net.TCPAddr) bool {
	if addr.IP.String() == "127.0.0.1" {
		return true
	}

	for i := range localAddrList {
		if localAddrList[i].Contains(addr.IP) {
			return true
		}
	}

	return false
}
