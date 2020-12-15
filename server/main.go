package main

import (
	"context"
	"flag"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"

	"github.com/coocood/freecache"
	socks5 "github.com/extrame/go-socks5"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"

	"github.com/extrame/grpc-socks/lib"
	"github.com/sirupsen/logrus"

	// "github.com/extrame/grpc-socks/log"
	"github.com/extrame/grpc-socks/pb"
)

var (
	remoteAddr   = ":50051"
	localAddr    = "127.0.0.1:50050"
	maxPerRemote = 3
	debug        = false
)

var (
	showVersion = false
	version     = "self-build"
	buildAt     = ""
	grpcHub     = &hub{
		serverToken: append([]byte(version), append([]byte("@"), []byte(buildAt)...)...),
		connected:   make(map[string]*client),
	}
)

func init() {
	flag.StringVar(&localAddr, "l", localAddr, "local addr")
	flag.StringVar(&remoteAddr, "r", remoteAddr, "remote addr")
	flag.IntVar(&maxPerRemote, "m", maxPerRemote, "max client per remote")
	flag.BoolVar(&debug, "d", debug, "debug mode")
	flag.BoolVar(&showVersion, "v", false, "show version then exit")

	flag.Parse()

	if showVersion {
		logrus.Infof("version:%s, build at %q using %s", version, buildAt, runtime.Version())
		os.Exit(0)
	}

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	encoding.RegisterCompressor(lib.Snappy())
}

func main() {
	ln, err := net.Listen("tcp", remoteAddr)
	if err != nil {
		logrus.Fatalf("failed to listen: %s", err)
	}
	defer ln.Close()

	logrus.Infof("starting proxy server at %q ...", remoteAddr)

	m := cmux.New(ln)

	httpL := m.Match(cmux.HTTP1Fast())
	httpS := &http.Server{
		Handler: nil,
	}
	go httpS.Serve(httpL)

	grpcL := m.Match(cmux.Any())
	grpcS := grpc.NewServer(grpc.Creds(lib.ServerTLS()), grpc.StreamInterceptor(interceptor))
	defer grpcS.GracefulStop()
	pb.RegisterProxyServer(grpcS, grpcHub)
	go func() {
		err := grpcS.Serve(grpcL)
		if err != nil {
			logrus.Fatalf("failed to serve grpc: %s", err.Error())
		} else {
			logrus.Info("start grpc...")
		}
	}()

	go func() {
		if err := m.Serve(); err != nil {
			logrus.Fatalf("failed to serve: %s", err)
		} else {
			logrus.Info("start http for grpc...")
		}
	}()

	// 从客户端迁移过来，将监听逻辑改到服务器端
	conf := &socks5.Config{
		Resolver: DNSResolver{cache: freecache.NewCache(100 * 1024 * 1024)},
		OnNewSession: func(ctx context.Context, addr net.Addr) context.Context {
			logrus.WithField("addr", addr.String()).Info("new session")
			ctxNew := context.WithValue(ctx, socks5.ClientID, addr.String())
			return ctxNew
		},
		Dial: DialFunc,
	}

	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}
	logrus.Infoln("start socks5 at ...", localAddr)
	if err := server.ListenAndServe("tcp", localAddr); err != nil {
		panic(err)
	}
}

func interceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, ss)
}
