package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRejectedUploadBodyIsDrainedBeforeResponseFinishes(t *testing.T) {
	consumed := make(chan int64, 1)
	inner := FinishUploadBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "forbidden")
	}))
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := &countedUploadBody{ReadCloser: r.Body}
		r.Body = body
		inner.ServeHTTP(w, r)
		consumed <- body.read
	}))
	defer server.Close()
	req, err := http.NewRequest(http.MethodPut, server.URL+"/admin-api/v1/system/files/1/content", bytes.NewReader(bytes.Repeat([]byte("x"), 1024*1024)))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := server.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil || resp.StatusCode != http.StatusForbidden || string(data) != "forbidden" {
		t.Fatal("rejection lost", resp.StatusCode, err)
	}
	if size := <-consumed; size != 1024*1024 {
		t.Fatal("body was not consumed", size)
	}
}

type countedUploadBody struct {
	io.ReadCloser
	read int64
}

func (b *countedUploadBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	b.read += int64(n)
	return n, err
}

func TestRejectedUploadDrainHasADeadline(t *testing.T) {
	finished := make(chan struct{})
	inner := finishUploadBody(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}), 50*time.Millisecond)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inner.ServeHTTP(w, r)
		close(finished)
	}))
	defer server.Close()
	conn, err := net.Dial("tcp", strings.TrimPrefix(server.URL, "http://"))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	_, err = fmt.Fprint(conn, "PUT /admin-api/v1/system/files/1/content HTTP/1.1\r\nHost: test\r\nContent-Length: 1048576\r\n\r\nx")
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("rejected upload waited indefinitely for its body")
	}
}
