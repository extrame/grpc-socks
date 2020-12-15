package main

import (
	"context"
	"net"

	"github.com/coocood/freecache"
	socks5 "github.com/extrame/go-socks5"

	"github.com/extrame/grpc-socks/pb"
)

type DNSResolver struct {
	cache *freecache.Cache
}

var expireSeconds = 7200

var nameCtxKey = struct{}{}

// DNSResolver uses the remote DNS to resolve host names
func (d DNSResolver) Resolve(ctx context.Context, name string) (context.Context, net.IP, error) {
	var clientId = ctx.Value(socks5.ClientID).(string)

	if v, err := d.cache.Get([]byte(name)); err == nil {
		return ctx, v, nil
	}

	var grpcClient = selectClient(clientId)

	ipResp, err := grpcClient.ResolveIP(ctx, &pb.IPAddr{
		Address: name,
	})

	if err == nil {
		d.cache.Set([]byte(name), ipResp.Data, expireSeconds)
	}

	return ctx, ipResp.Data, err
}
