package storage

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLocalAtomicImmutableAndBounds(t *testing.T) {
	for _, key := range []string{strings.Repeat("a", 32), "20261030/" + strings.Repeat("b", 32)} {
		t.Run(key, func(t *testing.T) { testLocalImmutableKey(t, key) })
	}
}

func testLocalImmutableKey(t *testing.T, key string) {
	t.Helper()
	s, err := New(Config{Driver: "local", Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
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
	for _, key := range []string{strings.Repeat("a", 32), "20261030/" + strings.Repeat("b", 32)} {
		t.Run(key, func(t *testing.T) { testS3ConditionalKey(t, key) })
	}
}

func testS3ConditionalKey(t *testing.T, key string) {
	t.Helper()
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/private/"+key {
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
	s, err := New(Config{Driver: "s3", Root: t.TempDir(), Endpoint: strings.TrimPrefix(server.URL, "http://"), Bucket: "private", AccessKey: "test-key", SecretKey: "test-secret", Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		err = s.Put(context.Background(), key, strings.NewReader("abc"), 3)
		if i == 0 && err != nil {
			t.Fatal(err)
		}
		if i == 1 && !errors.Is(err, ErrExists) {
			t.Fatal("conditional conflict not mapped", err)
		}
	}
}

func TestDatedKeysRejectInvalidDatesAndPathEscapes(t *testing.T) {
	name := strings.Repeat("a", 32)
	for _, key := range []string{name, "20240229/" + name, "20261030/" + name} {
		if !validKey(key) {
			t.Errorf("valid key rejected: %q", key)
		}
	}
	s, err := New(Config{Driver: "local", Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"20260229/" + name, "20260431/" + name, "00000101/" + name,
		"20261301/" + name, "2026/1/01/" + name, "/20261030/" + name,
		"2026/10/../" + name, "20261030/../../" + name, "2026\\10\\30\\" + name,
		"2026/10/30/" + name, "20261030\\" + name, "20261030/../" + name,
		"20261030/" + strings.ToUpper(name), "20261030/" + name + ".png",
	} {
		if validKey(key) {
			t.Errorf("unsafe key accepted: %q", key)
		}
		if err = s.Put(context.Background(), key, strings.NewReader("x"), 1); err == nil {
			t.Errorf("unsafe write accepted: %q", key)
		}
		if _, err = s.Open(context.Background(), key); err == nil {
			t.Errorf("unsafe read accepted: %q", key)
		}
		if err = s.Delete(context.Background(), key); err == nil {
			t.Errorf("unsafe delete accepted: %q", key)
		}
	}
}

func TestDatedDirectoriesCannotEscapeViaSymlink(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	s, err := New(Config{Driver: "local", Root: root})
	if err != nil {
		t.Fatal(err)
	}
	if err = os.MkdirAll(outside, 0700); err != nil {
		t.Fatal(err)
	}
	name := strings.Repeat("a", 32)
	protected := filepath.Join(outside, name)
	if err = os.WriteFile(protected, []byte("private"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Symlink(outside, filepath.Join(root, "20261030")); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err = s.Put(ctx, "20261030/"+strings.Repeat("b", 32), strings.NewReader("x"), 1); err == nil {
		t.Fatal("write escaped storage root")
	}
	if _, err = s.Open(ctx, "20261030/"+name); err == nil {
		t.Fatal("read escaped storage root")
	}
	if err = s.Delete(ctx, "20261030/"+name); err == nil {
		t.Fatal("delete escaped storage root")
	}
	data, err := os.ReadFile(protected)
	if err != nil || string(data) != "private" {
		t.Fatal("outside file changed", err)
	}
	if _, err = os.Stat(filepath.Join(outside, strings.Repeat("b", 32))); !os.IsNotExist(err) {
		t.Fatal("outside file created", err)
	}
}

func TestS3ProducerCleansItsConfiguredSpool(t *testing.T) {
	root := t.TempDir()
	store, err := New(Config{Driver: "s3", Root: root, Endpoint: "127.0.0.1:1", Bucket: "test", AccessKey: "test-key", SecretKey: "test-secret", Region: "us-east-1"})
	if err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(root, ".spool")
	old := filepath.Join(dir, ".goseed-upload-old")
	recent := filepath.Join(dir, ".goseed-upload-recent")
	other := filepath.Join(dir, "unrelated")
	for _, path := range []string{old, recent, other} {
		if err = os.WriteFile(path, []byte("test"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	past := time.Now().Add(-48 * time.Hour)
	if err = os.Chtimes(old, past, past); err != nil {
		t.Fatal(err)
	}
	if err = os.Chtimes(other, past, past); err != nil {
		t.Fatal(err)
	}
	stop := Maintain(store)
	defer stop()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if _, err = os.Stat(old); os.IsNotExist(err) {
			break
		}
		time.Sleep(time.Millisecond)
	}
	if _, err = os.Stat(old); !os.IsNotExist(err) {
		t.Fatal("stale API spool was not removed")
	}
	for _, path := range []string{recent, other} {
		if _, err = os.Stat(path); err != nil {
			t.Fatal("unrelated/recent file removed", err)
		}
	}
}
