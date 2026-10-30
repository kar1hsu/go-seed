package svc

import (
	"errors"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/config"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/middleware"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/pkg/storage"
	memberclient "github.com/kar1hsu/go-seed/services/member/rpc/member"
	systemclient "github.com/kar1hsu/go-seed/services/system/rpc/system"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Member        memberclient.Member
	memberClient  zrpc.Client
	Objects       storage.Store
	Config        config.Config
	BrowserOrigin rest.Middleware
	AdminAccess   rest.Middleware
	System        systemclient.System
	client        zrpc.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	objects, err := storage.New(c.Storage)
	if err != nil {
		return nil, err
	}
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
	memberGuard, err := rpcx.Client(c.MemberCredential)
	if err != nil {
		client.Conn().Close()
		return nil, err
	}
	c.MemberRPC.Middlewares.Duration = false
	memberClient, err := zrpc.NewClient(c.MemberRPC, zrpc.WithUnaryClientInterceptor(memberGuard))
	if err != nil {
		client.Conn().Close()
		return nil, err
	}
	return &ServiceContext{Member: memberclient.NewMember(memberClient), memberClient: memberClient, Objects: objects, Config: c, System: systemclient.NewSystem(client), client: client, BrowserOrigin: origin.Handle, AdminAccess: middleware.NewAdminAccessMiddleware(tokens).Handle}, nil
}
func (s *ServiceContext) Close() error {
	return errors.Join(s.client.Conn().Close(), s.memberClient.Conn().Close())
}
