package rpcx

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"os"
)

type ClientTLS struct {
	CAFile     string `json:",optional"`
	ServerName string `json:",optional"`
}
type ServerTLS struct {
	CertFile string `json:",optional"`
	KeyFile  string `json:",optional"`
}

func (c ClientTLS) Options(mode string) ([]zrpc.ClientOption, error) {
	if c.CAFile == "" && c.ServerName == "" {
		if mode == "pro" {
			return nil, errors.New("production RPC requires TLS CA and server name")
		}
		return nil, nil
	}
	if c.CAFile == "" || c.ServerName == "" {
		return nil, errors.New("incomplete RPC client TLS configuration")
	}
	pem, err := os.ReadFile(c.CAFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pem) {
		return nil, errors.New("invalid RPC CA")
	}
	return []zrpc.ClientOption{zrpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12, RootCAs: pool, ServerName: c.ServerName}))}, nil
}
func (c ServerTLS) Options(mode string) ([]grpc.ServerOption, error) {
	if c.CertFile == "" && c.KeyFile == "" {
		if mode == "pro" {
			return nil, errors.New("production RPC requires TLS certificate and key")
		}
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, err
	}
	return []grpc.ServerOption{grpc.Creds(credentials.NewTLS(&tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{cert}}))}, nil
}
