package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kar1hsu/go-seed/services/system/rpc/internal/config"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/server"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/system.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSystemServer(grpcServer, server.NewSystemServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
