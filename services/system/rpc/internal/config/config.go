package config

import (
	"time"

	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL          dbx.MySQLConfig
	Cache          redis.RedisConf
	Tokens         identity.Config
	RefreshTTL     time.Duration `json:",default=168h"`
	AdminAPISecret string
	Version        string `json:",default=0.1.0"`
}
