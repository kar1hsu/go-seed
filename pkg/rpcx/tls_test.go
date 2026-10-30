package rpcx

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	pb "google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestTLSRequiredAndVerified(t *testing.T) {
	if _, err := (ClientTLS{}).Options("pro"); err == nil {
		t.Fatal("plaintext client allowed")
	}
	if _, err := (ServerTLS{}).Options("pro"); err == nil {
		t.Fatal("plaintext server allowed")
	}
	source := httptest.NewTLSServer(nil)
	cert := source.TLS.Certificates[0]
	source.Close()
	dir := t.TempDir()
	certFile, keyFile := filepath.Join(dir, "cert.pem"), filepath.Join(dir, "key.pem")
	key, err := x509.MarshalPKCS8PrivateKey(cert.PrivateKey)
	if err != nil {
		t.Fatal(err)
	}
	os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]}), 0600)
	os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: key}), 0600)
	opts, err := (ServerTLS{certFile, keyFile}).Options("pro")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer(opts...)
	pb.RegisterHealthServer(server, health.NewServer())
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(listener)
	defer server.Stop()
	for _, name := range []string{"example.com", "wrong.example"} {
		opts, err := (ClientTLS{certFile, name}).Options("pro")
		if err != nil {
			t.Fatal(err)
		}
		client, err := zrpc.NewClient(zrpc.RpcClientConf{Endpoints: []string{listener.Addr().String()}, NonBlock: true}, opts...)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		_, err = pb.NewHealthClient(client.Conn()).Check(ctx, &pb.HealthCheckRequest{})
		cancel()
		client.Conn().Close()
		if (err == nil) != (name == "example.com") {
			t.Fatalf("TLS hostname verification %s: %v", name, err)
		}
	}
}
