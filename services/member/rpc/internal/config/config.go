package config

import (
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"time"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL             dbx.MySQLConfig
	Cache             redis.RedisConf
	Tokens            identity.Config
	RefreshTTL        time.Duration `json:",default=168h"`
	AdminAPISecret    string
	AppAPISecret      string
	SystemRPC         zrpc.RpcClientConf
	SystemCredential  rpcx.Credential
	RegisterPerMinute int `json:",default=10,range=[1:10000]"`
	LoginPerMinute    int `json:",default=120,range=[1:10000]"`
}
