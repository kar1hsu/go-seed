package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/kar1hsu/go-seed/pkg/rpcx"
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

	methods := []string{
		pb.System_ListAuditLogs_FullMethodName,
		pb.System_GetAuditLog_FullMethodName,
		pb.System_ListAdministrators_FullMethodName,
		pb.System_CreateAdministrator_FullMethodName,
		pb.System_UpdateAdministrator_FullMethodName,
		pb.System_AssignAdministratorRoles_FullMethodName,
		pb.System_ListRoles_FullMethodName,
		pb.System_CreateRole_FullMethodName,
		pb.System_UpdateRole_FullMethodName,
		pb.System_GrantRole_FullMethodName,
		pb.System_DeleteRole_FullMethodName,
		pb.System_ListPermissions_FullMethodName,

		pb.System_AdminLogin_FullMethodName,
		pb.System_AdminRefresh_FullMethodName,
		pb.System_AdminLogout_FullMethodName,
		pb.System_AdminCurrent_FullMethodName,
		pb.System_AdminChangePassword_FullMethodName,
	}
	callers := []rpcx.Caller{{Name: "admin-api", Secret: c.AdminAPISecret, Methods: methods}}
	public := []string{
		pb.System_Ping_FullMethodName,
		pb.System_CheckReady_FullMethodName,
		"/grpc.health.v1.Health/Check",
	}
	guard, err := rpcx.Server(callers, public)
	if err != nil {
		log.Fatal(err)
	}
	c.Middlewares.StatConf.IgnoreContentMethods = append(c.Middlewares.StatConf.IgnoreContentMethods, methods...)
	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSystemServer(grpcServer, server.NewSystemServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()
	s.AddUnaryInterceptors(guard)

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
