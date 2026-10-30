package config

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	MemberRPC        zrpc.RpcClientConf
	MemberCredential rpcx.Credential
	rest.RestConf
	Tokens         identity.Config
	CookieSecure   bool
	AllowedOrigins string
	SystemRPC      zrpc.RpcClientConf
}
