// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/clientip"
	"github.com/kar1hsu/go-seed/pkg/timezone"
	"log"

	"github.com/kar1hsu/go-seed/apps/app/api/internal/config"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/handler"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/pkg/xhttp"
	memberpb "github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/app-api.yaml", "the config file")

func main() {
	if err := timezone.Configure(); err != nil {
		log.Fatal(err)
	}
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	httpx.SetErrorHandlerCtx(xhttp.ErrorHandler)

	c.Middlewares.Log = false // Never log authentication request bodies.
	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		log.Fatal(err)
	}
	defer ctx.Close()
	xhttp.RegisterHealth(server, func(callCtx context.Context) error {
		if _, err := ctx.System.CheckReady(callCtx, &pb.CheckReadyRequest{}); err != nil {
			return err
		}
		_, err := ctx.Member.CheckReady(callCtx, &memberpb.Empty{})
		return err
	})
	ipMiddleware, err := clientip.Middleware(c.TrustedProxies)
	if err != nil {
		log.Fatal(err)
	}
	server.Use(ipMiddleware)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
