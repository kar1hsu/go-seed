package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

func TestLocalAtomicImmutableAndBounds(t *testing.T) {
	s, err := New(Config{Driver: "local", Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	key := strings.Repeat("a", 32)
	if err = s.Put(ctx, key, strings.NewReader("partial"), 10); !errors.Is(err, ErrSize) {
		t.Fatal(err)
	}
	if _, err = s.Open(ctx, key); err == nil {
		t.Fatal("partial object published")
	}
	var success atomic.Int32
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.Put(ctx, key, bytes.NewReader([]byte("content")), 7)
			if err == nil {
				success.Add(1)
			} else if !errors.Is(err, ErrExists) {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatal("multiple immutable writers won")
	}
	r, err := s.Open(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := io.ReadAll(r)
	r.Close()
	if string(b) != "content" {
		t.Fatal("bad bytes")
	}
	if err = s.Put(ctx, "../outside", strings.NewReader("x"), 1); err == nil {
		t.Fatal("path escape accepted")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if err = s.Put(cancelled, strings.Repeat("b", 32), strings.NewReader("x"), 1); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if err = s.Delete(ctx, key); err != nil {
		t.Fatal("delete not idempotent", err)
	}
}
func TestS3ConditionalWrite(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/private/"+strings.Repeat("a", 32) {
			t.Error("unexpected S3 request", r.Method, r.URL.Path)
			w.WriteHeader(400)
			return
		}
		if r.Header.Get("If-None-Match") != "*" {
			t.Error("S3 must reject overwrite")
		}
		if !strings.Contains(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256") {
			t.Error("missing signature")
		}
		if calls.Add(1) > 1 {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(412)
			io.WriteString(w, "<Error><Code>PreconditionFailed</Code></Error>")
			return
		}
		w.Header().Set("ETag", `"900150983cd24fb0d6963f7d28e17f72"`)
	}))
	defer server.Close()
	s, err := New(Config{Driver: "s3", Endpoint: strings.TrimPrefix(server.URL, "http://"), Bucket: "private", AccessKey: "test-key", SecretKey: "test-secret", Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		err = s.Put(context.Background(), strings.Repeat("a", 32), strings.NewReader("abc"), 3)
		if i == 0 && err != nil {
			t.Fatal(err)
		}
		if i == 1 && !errors.Is(err, ErrExists) {
			t.Fatal("conditional conflict not mapped", err)
		}
	}
}
