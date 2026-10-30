package config

import (
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type DomainCaller struct {
	rpcx.Credential
	PermissionPrefix string
}
type Config struct {
	TLS           rpcx.ServerTLS `json:",optional"`
	DomainCallers []DomainCaller `json:",optional"`
	Storage       storage.Config
	zrpc.RpcServerConf
	MySQL          dbx.MySQLConfig
	Cache          redis.RedisConf
	Tokens         identity.SignerConfig
	RefreshTTL     time.Duration `json:",default=168h"`
	AdminAPISecret string
	Version        string `json:",default=0.1.0"`
}
