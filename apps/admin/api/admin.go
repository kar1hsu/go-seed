// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/xhttp"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/rest/httpx"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/config"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/handler"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/admin-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	httpx.SetErrorHandlerCtx(xhttp.ErrorHandler)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	defer ctx.Close()
	xhttp.RegisterHealth(server, func(callCtx context.Context) error {
		_, err := ctx.System.CheckReady(callCtx, &pb.CheckReadyRequest{})
		return err
	})
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
