package xhttp

import (
	"bufio"
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"github.com/zeromicro/go-zero/rest/router"
)

type accessRecordKey struct{}

type accessRecord struct {
	route string // Set by Handle before the go-zero chain starts.
	code  atomic.Int64
}

type accessWriter struct {
	http.ResponseWriter
	status atomic.Int64
	bytes  atomic.Int64
}

func (w *accessWriter) WriteHeader(code int) {
	w.status.CompareAndSwap(0, int64(code))
	w.ResponseWriter.WriteHeader(code)
}

func (w *accessWriter) Write(b []byte) (int, error) {
	w.status.CompareAndSwap(0, http.StatusOK)
	n, err := w.ResponseWriter.Write(b)
	w.bytes.Add(int64(n))
	return n, err
}

// Unwrap keeps http.ResponseController working for streamed downloads.
func (w *accessWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *accessWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *accessWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, errors.New("hijack unsupported")
}

type accessRouter struct {
	httpx.Router
	service string
}

// NewAccessRouter wraps the whole go-zero chain, so it sees timeout, panic,
// body-limit and unmatched responses that server.Use middleware cannot observe.
func NewAccessRouter(service string) httpx.Router {
	return &accessRouter{Router: router.NewRouter(), service: service}
}

func (a *accessRouter) Handle(method, path string, handler http.Handler) error {
	return a.Router.Handle(method, path, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if record, ok := r.Context().Value(accessRecordKey{}).(*accessRecord); ok {
			record.route = path
		}
		handler.ServeHTTP(w, r)
	}))
}

func (a *accessRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	started := time.Now()
	// go-zero logs and spans read RequestURI; handlers read r.URL.
	r.RequestURI = r.URL.Path
	// Trace context is only accepted from inside the system.
	r.Header.Del("traceparent")
	r.Header.Del("tracestate")
	r.Header.Del("baggage")
	record := &accessRecord{route: "unmatched"}
	writer := &accessWriter{ResponseWriter: w}
	defer func() { a.log(r, w, writer, record, time.Since(started)) }()
	a.Router.ServeHTTP(writer, r.WithContext(context.WithValue(r.Context(), accessRecordKey{}, record)))
}

var accessQueryValue = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,64}$`)

func (a *accessRouter) log(r *http.Request, w http.ResponseWriter, writer *accessWriter, record *accessRecord, elapsed time.Duration) {
	if record.route == "/healthz" || record.route == "/readyz" {
		return
	}
	status := writer.status.Load()
	if status == 0 {
		status = http.StatusOK
	}
	fields := []logx.LogField{
		logx.Field("service", a.service),
		logx.Field("method", accessMethod(r.Method)),
		logx.Field("route", record.route),
		logx.Field("status", status),
		logx.Field("bytes", writer.bytes.Load()),
	}
	if r.ContentLength > 0 {
		fields = append(fields, logx.Field("requestBytes", r.ContentLength))
	}
	if code := record.code.Load(); code != 0 {
		fields = append(fields, logx.Field("businessCode", code))
	}
	if parts := strings.Split(w.Header().Get("Traceparent"), "-"); len(parts) == 4 {
		fields = append(fields, logx.Field("trace", parts[1]))
	}
	if query := accessQuery(r.URL.RawQuery); len(query) > 0 {
		fields = append(fields, logx.Field("query", query))
	}
	logger := logx.WithDuration(elapsed)
	if status >= http.StatusInternalServerError {
		logger.Errorw("http access", fields...)
		return
	}
	logger.Infow("http access", fields...)
}

func accessQuery(raw string) map[string]string {
	if raw == "" {
		return nil
	}
	values, _ := url.ParseQuery(raw)
	kept := map[string]string{}
	for _, key := range []string{"page", "pageSize", "enabled", "status"} {
		if value := values.Get(key); accessQueryValue.MatchString(value) {
			kept[key] = value
		}
	}
	return kept
}

// accessMethod bounds a client-controlled token before it reaches the log.
func accessMethod(method string) string {
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return method
	}
	return "OTHER"
}
