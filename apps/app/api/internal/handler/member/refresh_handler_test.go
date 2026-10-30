package member

import (
	"context"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/pkg/xhttp"
	client "github.com/kar1hsu/go-seed/services/member/rpc/member"
	"github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"net/http/httptest"
	"strings"
	"testing"
)

type refreshStub struct {
	client.Member
	key    string
	called bool
}

func (s *refreshStub) Refresh(ctx context.Context, _ *pb.RefreshRequest, _ ...grpc.CallOption) (*pb.SessionResponse, error) {
	s.called = true
	s.key = identity.RefreshKey(ctx)
	return nil, xerr.New(codes.Unavailable)
}
func TestRefreshHeaderOnlyContract(t *testing.T) {
	httpx.SetErrorHandlerCtx(xhttp.ErrorHandler)
	key, _ := identity.Random()
	for _, body := range []string{"", "null", "{}", `{"RetryKey":"untrusted-body"}`} {
		for _, header := range []string{key, "", "invalid"} {
			stub := &refreshStub{}
			r := httptest.NewRequest("POST", "/refresh", strings.NewReader(body))
			r.Header.Set("Content-Type", "application/json")
			if header != "" {
				r.Header.Set("X-GoSeed-Refresh-Key", header)
			}
			w := httptest.NewRecorder()
			MemberRefreshHandler(&svc.ServiceContext{Member: stub})(w, r)
			if header == "invalid" {
				if stub.called || w.Code != 400 {
					t.Fatal("invalid key reached RPC")
				}
			} else if !stub.called || stub.key != header || w.Code != 503 {
				t.Fatalf("header contract failed for body %q: status %d called %v", body, w.Code, stub.called)
			}
		}
	}
}
