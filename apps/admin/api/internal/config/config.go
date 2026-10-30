package config

import (
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/pkg/storage"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	SystemTLS        rpcx.ClientTLS `json:",optional"`
	MemberTLS        rpcx.ClientTLS `json:",optional"`
	TrustedProxies   string         `json:",optional"`
	MemberRPC        zrpc.RpcClientConf
	MemberCredential rpcx.Credential
	Storage          storage.Config
	rest.RestConf
	SystemRPC      zrpc.RpcClientConf
	Tokens         identity.Config
	RPCCredential  rpcx.Credential
	CookieSecure   bool
	AllowedOrigins string
}
