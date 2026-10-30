package main

import (
	"flag"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/config"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/server"
	"github.com/kar1hsu/go-seed/services/member/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"log"
)

func main() {
	file := flag.String("f", "etc/member.yaml", "configuration file")
	flag.Parse()
	var c config.Config
	conf.MustLoad(*file, &c, conf.UseEnv())
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()
	admin := []string{pb.Member_ListUsers_FullMethodName, pb.Member_GetUser_FullMethodName, pb.Member_SetUserStatus_FullMethodName}
	app := []string{pb.Member_Register_FullMethodName, pb.Member_Login_FullMethodName, pb.Member_Refresh_FullMethodName, pb.Member_Logout_FullMethodName, pb.Member_Current_FullMethodName, pb.Member_UpdateProfile_FullMethodName, pb.Member_ChangePassword_FullMethodName}
	guard, err := rpcx.Server([]rpcx.Caller{{Name: "admin-api", Secret: c.AdminAPISecret, Methods: admin}, {Name: "app-api", Secret: c.AppAPISecret, Methods: app}}, []string{pb.Member_CheckReady_FullMethodName, "/grpc.health.v1.Health/Check"})
	if err != nil {
		log.Fatal(err)
	}
	c.Middlewares.StatConf.IgnoreContentMethods = append(append(c.Middlewares.StatConf.IgnoreContentMethods, admin...), app...)
	s := zrpc.MustNewServer(c.RpcServerConf, func(server *grpc.Server) { pb.RegisterMemberServer(server, serverAdapter(ctx)) })
	defer s.Stop()
	s.AddUnaryInterceptors(guard)
	fmt.Printf("Starting member RPC at %s...\n", c.ListenOn)
	s.Start()
}
func serverAdapter(ctx *svc.ServiceContext) pb.MemberServer { return server.NewMemberServer(ctx) }
