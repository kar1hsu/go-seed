package svc

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/config"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/middleware"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	systemclient "github.com/kar1hsu/go-seed/services/system/rpc/system"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config        config.Config
	BrowserOrigin rest.Middleware
	AdminAccess   rest.Middleware
	System        systemclient.System
	client        zrpc.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	tokens, err := identity.New(c.Tokens)
	if err != nil {
		return nil, err
	}
	origin, err := middleware.NewBrowserOriginMiddleware(c.AllowedOrigins)
	if err != nil {
		return nil, err
	}
	interceptor, err := rpcx.Client(c.RPCCredential)
	if err != nil {
		return nil, err
	}
	// go-zero's duration interceptor logs request/response content on errors and slow calls.
	c.SystemRPC.Middlewares.Duration = false
	client, err := zrpc.NewClient(c.SystemRPC, zrpc.WithUnaryClientInterceptor(interceptor))
	if err != nil {
		return nil, err
	}
	return &ServiceContext{Config: c, System: systemclient.NewSystem(client), client: client, BrowserOrigin: origin.Handle, AdminAccess: middleware.NewAdminAccessMiddleware(tokens).Handle}, nil
}
func (s *ServiceContext) Close() error { return s.client.Conn().Close() }
