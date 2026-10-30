// Probe is the container healthcheck client for GoSeed processes.
package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/kar1hsu/go-seed/pkg/timezone"
	"net/http"
	"os"
	"time"

	"github.com/kar1hsu/go-seed/pkg/rpcx"
	memberpb "github.com/kar1hsu/go-seed/services/member/rpc/pb"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc/health/grpc_health_v1"
	"strings"
)

func main() {
	if err := timezone.Configure(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	target := flag.String("rpc", "", "RPC target")
	url := flag.String("http", "", "HTTP health endpoint")
	domain := flag.String("domain", "system", "RPC domain: system or member")
	ready := flag.Bool("ready", false, "check RPC dependencies as well")
	flag.Parse()
	if err := check(*target, *url, *ready, *domain); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func check(target, url string, ready bool, domain string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP health: %d", resp.StatusCode)
		}
		return nil
	}
	if target == "" {
		return fmt.Errorf("pass -rpc or -http")
	}
	tlsOptions, err := (rpcx.ClientTLS{CAFile: os.Getenv("RPC_TLS_CA"), ServerName: os.Getenv(strings.ToUpper(domain) + "_RPC_TLS_NAME")}).Options(os.Getenv("GOSEED_MODE"))
	if err != nil {
		return err
	}
	client, err := zrpc.NewClient(zrpc.RpcClientConf{Endpoints: []string{target}, NonBlock: true}, tlsOptions...)
	if err != nil {
		return err
	}
	conn := client.Conn()
	defer conn.Close()
	result, err := grpc_health_v1.NewHealthClient(conn).Check(ctx, &grpc_health_v1.HealthCheckRequest{})
	if err != nil {
		return err
	}
	if result.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("RPC is not serving")
	}
	if ready {
		switch domain {
		case "system":
			_, err = pb.NewSystemClient(conn).CheckReady(ctx, &pb.CheckReadyRequest{})
		case "member":
			_, err = memberpb.NewMemberClient(conn).CheckReady(ctx, &memberpb.Empty{})
		default:
			return fmt.Errorf("unknown RPC domain")
		}
	}
	return err
}
