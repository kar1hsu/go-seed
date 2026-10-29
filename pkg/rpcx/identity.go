package rpcx

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
)

type Credential struct {
	Name   string
	Secret string
}
type Caller struct {
	Name    string
	Secret  string
	Methods []string
}

func Client(c Credential) (grpc.UnaryClientInterceptor, error) {
	if c.Name == "" || len(c.Secret) < 32 {
		return nil, errors.New("invalid RPC credential")
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoke grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		md = md.Copy()
		md.Set("x-service-name", c.Name)
		md.Set("x-service-secret", c.Secret)
		md.Delete("x-user-token")
		if token := identity.RawToken(ctx); token != "" {
			md.Set("x-user-token", token)
		}
		return invoke(metadata.NewOutgoingContext(ctx, md), method, req, reply, cc, opts...)
	}, nil
}

// Server rejects all non-public methods unless the configured service has an explicit grant.
func Server(callers []Caller, public []string) (grpc.UnaryServerInterceptor, error) {
	type grant struct {
		secret  string
		methods map[string]bool
	}
	allowed := make(map[string]grant)
	open := make(map[string]bool)
	for _, method := range public {
		open[method] = true
	}
	for _, caller := range callers {
		if caller.Name == "" || len(caller.Secret) < 32 {
			return nil, errors.New("invalid RPC caller")
		}
		if _, exists := allowed[caller.Name]; exists {
			return nil, errors.New("duplicate RPC caller")
		}
		methods := make(map[string]bool)
		for _, method := range caller.Methods {
			methods[method] = true
		}
		allowed[caller.Name] = grant{caller.Secret, methods}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, next grpc.UnaryHandler) (any, error) {
		if open[info.FullMethod] {
			return next(ctx, req)
		}
		md, _ := metadata.FromIncomingContext(ctx)
		names, secrets := md.Get("x-service-name"), md.Get("x-service-secret")
		if len(names) != 1 || len(secrets) != 1 {
			return nil, xerr.New(codes.Unauthenticated)
		}
		caller, ok := allowed[names[0]]
		if !ok || subtle.ConstantTimeCompare([]byte(secrets[0]), []byte(caller.secret)) != 1 {
			return nil, xerr.New(codes.Unauthenticated)
		}
		if !caller.methods[info.FullMethod] {
			return nil, xerr.New(codes.PermissionDenied)
		}
		values := md.Get("x-user-token")
		if len(values) > 1 {
			return nil, xerr.New(codes.Unauthenticated)
		}
		if len(values) == 1 {
			ctx = identity.WithToken(ctx, values[0])
		}
		return next(ctx, req)
	}, nil
}
