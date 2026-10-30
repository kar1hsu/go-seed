package rpcx

import (
	"context"
	"strings"
	"testing"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func TestServiceMethodGrants(t *testing.T) {
	secret := strings.Repeat("s", 48)
	guard, err := Server([]Caller{{Name: "admin-api", Secret: secret, Methods: []string{"/System/Me"}}}, []string{"/System/Ping"})
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, method, service, key string
		code                       codes.Code
	}{
		{"public", "/System/Ping", "", "", codes.OK},
		{"no service", "/System/Me", "", "", codes.Unauthenticated},
		{"wrong key", "/System/Me", "admin-api", "wrong", codes.Unauthenticated},
		{"wrong service", "/System/Me", "app-api", secret, codes.Unauthenticated},
		{"ungranted", "/System/AdminDelete", "admin-api", secret, codes.PermissionDenied},
		{"valid", "/System/Me", "admin-api", secret, codes.OK},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-name", tt.service, "x-service-secret", tt.key, "x-user-token", "signed-user-token"))
			called := false
			_, err := guard(ctx, nil, &grpc.UnaryServerInfo{FullMethod: tt.method}, func(ctx context.Context, _ any) (any, error) {
				called = true
				if tt.name == "valid" && identity.RawToken(ctx) != "signed-user-token" {
					t.Fatal("token lost")
				}
				return nil, nil
			})
			if status.Code(err) != tt.code || called != (tt.code == codes.OK) {
				t.Fatalf("code=%v called=%v", status.Code(err), called)
			}
		})
	}
}
func TestClientReplacesUntrustedMetadata(t *testing.T) {
	interceptor, _ := Client(Credential{Name: "admin-api", Secret: strings.Repeat("s", 48)})
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("x-user-token", "spoofed", "x-service-name", "spoofed", "x-client-ip", "1.2.3.4", "x-refresh-key", "spoofed"))
	err := interceptor(ctx, "/System/Me", nil, nil, nil, func(ctx context.Context, _ string, _, _ any, _ *grpc.ClientConn, _ ...grpc.CallOption) error {
		md, _ := metadata.FromOutgoingContext(ctx)
		if len(md.Get("x-client-ip")) != 0 || len(md.Get("x-refresh-key")) != 0 || len(md.Get("x-user-token")) != 0 || len(md.Get("x-service-name")) != 1 || md.Get("x-service-name")[0] != "admin-api" {
			t.Fatal("untrusted metadata survived")
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
