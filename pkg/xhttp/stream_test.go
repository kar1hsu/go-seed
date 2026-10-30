package xhttp

import (
	"context"
	"github.com/zeromicro/go-zero/rest/handler"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDownloadFlushesBeforeStorageFinishes(t *testing.T) {
	reader, writer := io.Pipe()
	defer reader.Close()
	defer writer.Close()
	done := make(chan error, 1)
	server := httptest.NewServer(handler.TraceHandler("stream", "/")(handler.TimeoutHandler(0)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), time.Second)
		defer cancel()
		deadline, _ := ctx.Deadline()
		done <- Stream(ctx, w, reader, deadline)
	}))))
	defer server.Close()
	first := make(chan error, 1)
	go func() {
		res, err := http.Get(server.URL)
		if err == nil {
			defer res.Body.Close()
			b := make([]byte, 5)
			_, err = io.ReadFull(res.Body, b)
			if string(b) != "hello" {
				err = io.ErrUnexpectedEOF
			}
		}
		first <- err
	}()
	if _, err := writer.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-first:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("first chunk buffered until EOF")
	}
	writer.Close()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("stream did not stop")
	}
}
func TestStreamCancellationClosesBlockedRead(t *testing.T) {
	r, w := io.Pipe()
	defer w.Close()
	done := make(chan error, 1)
	server := httptest.NewServer(http.HandlerFunc(func(out http.ResponseWriter, req *http.Request) {
		ctx, cancel := context.WithTimeout(req.Context(), 30*time.Millisecond)
		defer cancel()
		deadline, _ := ctx.Deadline()
		done <- Stream(ctx, out, r, deadline)
	}))
	defer server.Close()
	go func() {
		res, err := http.Get(server.URL)
		if err == nil {
			res.Body.Close()
		}
	}()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("cancellation lost")
		}
	case <-time.After(time.Second):
		t.Fatal("blocked storage read leaked")
	}
}
func TestProductionCookieValidation(t *testing.T) {
	for _, tc := range []struct {
		mode    string
		secure  bool
		origins string
		ok      bool
	}{{"dev", false, "http://localhost", true}, {"pro", false, "https://example.test", false}, {"pro", true, "http://example.test", false}, {"pro", true, "", false}, {"pro", true, "https://example.test", true}} {
		if err := ValidateProduction(tc.mode, tc.secure, tc.origins); (err == nil) != tc.ok {
			t.Fatalf("%+v: %v", tc, err)
		}
	}
}
