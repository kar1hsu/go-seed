package config

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	SystemRPC      zrpc.RpcClientConf
	Tokens         identity.Config
	RPCCredential  rpcx.Credential
	CookieSecure   bool
	AllowedOrigins string
}
