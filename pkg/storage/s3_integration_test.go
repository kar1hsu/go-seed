//go:build integration

package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

func TestS3RoundTrip(t *testing.T) {
	endpoint := os.Getenv("GOSEED_S3_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("set GOSEED_S3_TEST_ENDPOINT for an isolated S3 service")
	}
	ctx := context.Background()
	access, secret := os.Getenv("GOSEED_S3_TEST_ACCESS"), os.Getenv("GOSEED_S3_TEST_SECRET")
	client, err := minio.New(endpoint, &minio.Options{Creds: credentials.NewStaticV4(access, secret, ""), Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	bucket := fmt.Sprintf("goseed-test-%d", time.Now().UnixNano())
	if err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{Region: "us-east-1"}); err != nil {
		t.Fatal(err)
	}
	defer client.RemoveBucket(ctx, bucket)
	s, err := New(Config{Driver: "s3", Root: t.TempDir(), Endpoint: endpoint, Bucket: bucket, AccessKey: access, SecretKey: secret, Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err = s.Check(ctx); err != nil {
		t.Fatal(err)
	}
	key := strings.Repeat("c", 32)
	defer s.Delete(ctx, key)
	if err = s.Put(ctx, key, strings.NewReader("hello"), 5); err != nil {
		t.Fatal(err)
	}
	if err = s.Put(ctx, key, strings.NewReader("other"), 5); !errors.Is(err, ErrExists) {
		t.Fatal("S3 overwrote private object", err)
	}
	r, err := s.Open(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	r.Close()
	if err != nil || string(data) != "hello" {
		t.Fatal("S3 roundtrip mismatch", err)
	}
	if err = s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Open(ctx, key); err == nil {
		t.Fatal("S3 object survived deletion")
	}
}
