package logic

import (
	"context"
	"github.com/kar1hsu/go-seed/pkg/rpcx"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/config"
	"github.com/kar1hsu/go-seed/services/system/rpc/internal/svc"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"strings"
	"testing"
)

func TestDomainAccessScopeUsesAuthenticatedCaller(t *testing.T) {
	secret := strings.Repeat("s", 48)
	guard, err := rpcx.Server([]rpcx.Caller{{Name: "member-rpc", Secret: secret, Methods: []string{pb.System_CheckAccess_FullMethodName}}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	service := &svc.ServiceContext{Config: config.Config{DomainCallers: []config.DomainCaller{{Credential: rpcx.Credential{Name: "member-rpc", Secret: secret}, PermissionPrefix: "member."}}}}
	for _, permission := range []string{"system.user.read", "memberx.user.read", "", "*.user.read"} {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-name", "member-rpc", "x-service-secret", secret))
		_, err := guard(ctx, nil, &grpc.UnaryServerInfo{FullMethod: pb.System_CheckAccess_FullMethodName}, func(ctx context.Context, _ any) (any, error) {
			return NewCheckAccessLogic(ctx, service).CheckAccess(&pb.CheckAccessRequest{Permission: permission})
		})
		if status.Code(err) != codes.PermissionDenied {
			t.Fatalf("scope %q: %v", permission, err)
		}
	}
	// Raw, unverified metadata is not a service identity.
	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("x-service-name", "member-rpc"))
	_, err = NewCheckAccessLogic(ctx, service).CheckAccess(&pb.CheckAccessRequest{Permission: "member.user.read"})
	if status.Code(err) != codes.PermissionDenied {
		t.Fatal(err)
	}
}
