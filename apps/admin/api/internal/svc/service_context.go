package svc

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/config"
	systemclient "github.com/kar1hsu/go-seed/services/system/rpc/system"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config
	System systemclient.System
	client zrpc.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	client := zrpc.MustNewClient(c.SystemRPC)
	return &ServiceContext{Config: c, System: systemclient.NewSystem(client), client: client}
}
func (s *ServiceContext) Close() error { return s.client.Conn().Close() }
