package config

import (
	"github.com/kar1hsu/go-seed/pkg/dbx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL   dbx.MySQLConfig
	Cache   redis.RedisConf
	Version string `json:",default=0.1.0"`
}
