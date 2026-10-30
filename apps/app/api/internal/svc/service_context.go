package svc

import (
	"errors"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/config"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/middleware"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/pkg/xhttp"
	memberclient "github.com/kar1hsu/go-seed/services/member/rpc/member"
	systemclient "github.com/kar1hsu/go-seed/services/system/rpc/system"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Member        memberclient.Member
	memberClient  zrpc.Client
	Config        config.Config
	BrowserOrigin rest.Middleware
	MemberAccess  rest.Middleware
	System        systemclient.System
	client        zrpc.Client
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	if err := xhttp.ValidateProduction(c.Mode, c.CookieSecure, c.AllowedOrigins); err != nil {
		return nil, err
	}
	systemOpts, err := c.SystemTLS.Options(c.Mode)
	if err != nil {
		return nil, err
	}
	memberOpts, err := c.MemberTLS.Options(c.Mode)
	if err != nil {
		return nil, err
	}
	tokens, err := identity.NewVerifier(c.Tokens)
	if err != nil {
		return nil, err
	}
	origin, err := middleware.NewBrowserOriginMiddleware(c.AllowedOrigins)
	if err != nil {
		return nil, err
	}
	// go-zero's duration interceptor logs request/response content on errors and slow calls.
	c.SystemRPC.Middlewares.Duration = false
	client, err := zrpc.NewClient(c.SystemRPC, systemOpts...)
	if err != nil {
		return nil, err
	}
	memberGuard, err := rpcx.Client(c.MemberCredential)
	if err != nil {
		client.Conn().Close()
		return nil, err
	}
	c.MemberRPC.Middlewares.Duration = false
	memberClient, err := zrpc.NewClient(c.MemberRPC, append(memberOpts, zrpc.WithUnaryClientInterceptor(memberGuard))...)
	if err != nil {
		client.Conn().Close()
		return nil, err
	}
	return &ServiceContext{Member: memberclient.NewMember(memberClient), memberClient: memberClient, Config: c, System: systemclient.NewSystem(client), client: client, BrowserOrigin: origin.Handle, MemberAccess: middleware.NewMemberAccessMiddleware(tokens).Handle}, nil
}
func (s *ServiceContext) Close() error {
	return errors.Join(s.client.Conn().Close(), s.memberClient.Conn().Close())
}
