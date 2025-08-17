package main

import (
	"flag"
	"fmt"

	"notice/rpc/internal/server"
	"notice/rpc/internal/svc"
	"notice/rpc/notice/notice"

	"github.com/Skysolderone/zero_core/config"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/notice.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.RpcConfig
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		notice.RegisterNoticeServiceServer(grpcServer, server.NewNoticeServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
