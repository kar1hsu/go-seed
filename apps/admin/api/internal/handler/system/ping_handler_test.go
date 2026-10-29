package system

import (
	"context"
	"encoding/json"
	"net"
	"net/http/httptest"
	"testing"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/pkg/xhttp"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type probeServer struct {
	pb.UnimplementedSystemServer
	fail bool
}

func (s *probeServer) Ping(context.Context, *pb.PingRequest) (*pb.PingResponse, error) {
	if s.fail {
		return nil, status.Error(codes.Unavailable, "internal host password=secret")
	}
	return &pb.PingResponse{Service: "system-rpc", Version: "0.1.0", Time: "2026-10-29T00:00:00Z"}, nil
}
func TestPingThroughRPC(t *testing.T) {
	httpx.SetErrorHandlerCtx(xhttp.ErrorHandler)
	for _, fail := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "dependency failure"}[fail], func(t *testing.T) {
			listener := bufconn.Listen(1024 * 1024)
			server := grpc.NewServer()
			pb.RegisterSystemServer(server, &probeServer{fail: fail})
			go func() { _ = server.Serve(listener) }()
			t.Cleanup(func() { server.Stop(); _ = listener.Close() })
			conn, err := grpc.NewClient("passthrough:///test", grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return listener.Dial() }))
			if err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/admin-api/v1/system/ping", nil)
			PingHandler(&svc.ServiceContext{System: pb.NewSystemClient(conn)})(recorder, request)
			var body struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
				Data    *struct {
					Service string `json:"service"`
				} `json:"data"`
				TraceID string `json:"traceId"`
			}
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if fail {
				if recorder.Code != 503 || body.Code != 100503 || body.Data != nil || body.Message != "service unavailable" {
					t.Fatalf("unsafe/incorrect error: %s", recorder.Body)
				}
			} else if recorder.Code != 200 || body.Code != 0 || body.Data == nil || body.Data.Service != "system-rpc" {
				t.Fatalf("wrong success envelope: %s", recorder.Body)
			}
		})
	}
}
