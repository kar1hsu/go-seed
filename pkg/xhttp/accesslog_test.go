package xhttp

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx/logtest"
	ztrace "github.com/zeromicro/go-zero/core/trace"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
)

func TestAccessLogWithServerChain(t *testing.T) {
	// go-zero sets its tracer provider once per process. Keep the file exporter
	// and log collector isolated, including when go test repeats or shuffles tests.
	if os.Getenv("GOSEED_ACCESS_TEST_CHILD") != "1" {
		cmd := exec.Command(os.Args[0], "-test.run=^TestAccessLogWithServerChain$")
		cmd.Env = append(os.Environ(), "GOSEED_ACCESS_TEST_CHILD=1")
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("HTTP chain test: %v\n%s", err, out)
		}
		return
	}
	const forged = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	var c rest.RestConf
	if err := conf.LoadFromJsonBytes([]byte(`{"Name":"test-api","Host":"127.0.0.1","Port":0,"Mode":"test"}`), &c); err != nil {
		t.Fatal(err)
	}
	c.Middlewares.Log = false
	c.Telemetry.Batcher = "file"
	c.Telemetry.Endpoint = filepath.Join(t.TempDir(), "traces.json")
	c.Telemetry.Sampler = 1
	server := rest.MustNewServer(c, WithAccessLog(c.Name))
	logs := logtest.NewCollector(t)
	httpx.SetErrorHandlerCtx(ErrorHandler)
	t.Cleanup(func() { httpx.SetErrorHandlerCtx(nil) })
	server.AddRoute(rest.Route{Method: "GET", Path: "/v1/items/:id", Handler: func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("keyword") != "OBS_SEARCH" || r.Header.Get("baggage") != "" || r.Header.Get("tracestate") != "" || r.Header.Get("traceparent") != "" {
			t.Error("query lost or untrusted trace headers retained")
		}
		if r.UserAgent() != "OBS_PRIVATE_UA" || !reflect.DeepEqual(r.Header.Values("X-Forwarded-For"), []string{"203.0.113.77", "192.0.2.24"}) {
			t.Error("application middleware lost original client headers")
		}
		io.WriteString(w, ztrace.TraceIDFromContext(r.Context()))
	}})
	server.AddRoute(rest.Route{Method: "GET", Path: "/error", Handler: func(w http.ResponseWriter, r *http.Request) {
		httpx.ErrorCtx(r.Context(), w, xerr.New(codes.Unavailable))
	}})
	release := make(chan struct{})
	defer close(release)
	server.AddRoute(rest.Route{Method: "GET", Path: "/timeout", Handler: func(http.ResponseWriter, *http.Request) { <-release }}, rest.WithTimeout(15*time.Millisecond))
	server.AddRoute(rest.Route{Method: "GET", Path: "/panic", Handler: func(http.ResponseWriter, *http.Request) { panic("fixture panic") }})
	server.AddRoute(rest.Route{Method: "POST", Path: "/limit", Handler: func(http.ResponseWriter, *http.Request) { t.Error("oversize request reached handler") }}, rest.WithMaxBytes(3))
	server.AddRoute(rest.Route{Method: "POST", Path: "/login", Handler: func(w http.ResponseWriter, r *http.Request) { io.WriteString(w, "OBS_RESPONSE_SECRET") }})
	server.AddRoute(rest.Route{Method: "GET", Path: "/readyz", Handler: func(http.ResponseWriter, *http.Request) {}})
	server.AddRoute(rest.Route{Method: "GET", Path: "/stream", Handler: func(w http.ResponseWriter, r *http.Request) {
		if err := Stream(r.Context(), w, io.NopCloser(strings.NewReader("stream-data")), time.Now().Add(time.Second)); err != nil {
			t.Error(err)
		}
	}}, rest.WithTimeout(0))
	// Build the actual rest.Server chains, then serve them on a kernel-assigned port.
	built, err := rest.NewServerless(server)
	if err != nil {
		t.Fatal(err)
	}
	finished := make(chan struct{}, 1)
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { built.Serve(w, r); finished <- struct{}{} }))
	defer host.Close()
	for _, tc := range []struct {
		method, path, route, body string
		status, code              int
	}{
		{"GET", "/v1/items/7?keyword=OBS_SEARCH&page=2&enabled=all&status=x%20y", "/v1/items/:id", "", 200, 0},
		{"GET", "/error", "/error", "", 503, 100503}, {"GET", "/timeout", "/timeout", "", 504, 100504}, {"GET", "/panic?keyword=OBS_SEARCH", "/panic", "", 500, 0},
		{"POST", "/limit", "/limit", "large", 413, 0}, {"GET", "/v1/nope/OBS_PATH_SECRET", "unmatched", "", 404, 0}, {"POST", "/stream", "unmatched", "", 405, 0},
		{"GET", "/stream", "/stream", "", 200, 0}, {"POST", "/login", "/login", "OBS_PASSWORD", 200, 0}, {"GET", "/readyz", "/readyz", "", 200, 0},
	} {
		t.Run(tc.route+tc.method, func(t *testing.T) {
			logs.Reset()
			req, err := http.NewRequest(tc.method, host.URL+tc.path, strings.NewReader(tc.body))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("User-Agent", "OBS_PRIVATE_UA")
			req.Header.Add("X-Forwarded-For", "203.0.113.77")
			req.Header.Add("X-Forwarded-For", "192.0.2.24")
			req.Header.Set("Authorization", "Bearer OBS_TOKEN")
			req.Header.Set("Cookie", "test=OBS_COOKIE")
			req.Header.Set("traceparent", "00-"+forged+"-bbbbbbbbbbbbbbbb-01")
			req.Header.Set("tracestate", "test=OBS_STATE")
			req.Header.Set("baggage", "tenant=OBS_BAGGAGE")
			res, err := host.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer res.Body.Close()
			b, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}
			<-finished
			if res.StatusCode != tc.status {
				t.Fatalf("HTTP status %d, want %d", res.StatusCode, tc.status)
			}
			if tc.status == 405 && res.Header.Get("Allow") != "GET" {
				t.Errorf("Allow=%q", res.Header.Get("Allow"))
			}
			raw := logs.String()
			assertNoAccessSecrets(t, raw)
			var entries []map[string]any
			for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
				if line == "" {
					continue
				}
				var row map[string]any
				if err := json.Unmarshal([]byte(line), &row); err != nil {
					t.Fatal(err)
				}
				if row["content"] == "http access" {
					entries = append(entries, row)
				}
			}
			if tc.route == "/readyz" {
				if len(entries) != 0 {
					t.Error("logged health probe")
				}
				return
			}
			if len(entries) != 1 {
				t.Fatalf("got %d access records", len(entries))
			}
			row := entries[0]
			if row["status"] != float64(tc.status) || row["route"] != tc.route || row["method"] != tc.method || row["bytes"] != float64(len(b)) || row["service"] != c.Name {
				t.Errorf("incorrect access record: %v", row)
			}
			if tc.status >= 500 && row["level"] != "error" {
				t.Error("missing error level")
			}
			if tc.code != 0 {
				var envelope ErrorResponse
				if err := json.Unmarshal(b, &envelope); err != nil || envelope.Code != tc.code || row["businessCode"] != float64(tc.code) {
					t.Errorf("incorrect business code: body=%s, log=%v", b, row)
				}
			}
			if tc.status == 413 && row["requestBytes"] != float64(len(tc.body)) {
				t.Error("missing content length")
			}
			if tc.route == "unmatched" && row["trace"] != nil {
				t.Error("unmatched request inherited trace")
			}
			if tc.route == "/stream" && string(b) != "stream-data" {
				t.Error("incomplete stream")
			}
			if tc.route == "/v1/items/:id" {
				if !reflect.DeepEqual(row["query"], map[string]any{"page": "2", "enabled": "all"}) {
					t.Errorf("query=%v", row["query"])
				}
				if len(b) != 32 || string(b) == forged || row["trace"] != string(b) || !strings.Contains(res.Header.Get("Traceparent"), "-"+string(b)+"-") {
					t.Error("trace mismatch")
				}
			}
		})
	}
	ztrace.StopAgent()
	data, err := os.ReadFile(c.Telemetry.Endpoint)
	if err != nil {
		t.Fatal(err)
	}
	assertNoAccessSecrets(t, string(data))
	dec := json.NewDecoder(strings.NewReader(string(data)))
	found := false
	for dec.More() {
		var span struct {
			Name       string
			Attributes []struct {
				Key   string
				Value struct{ Value any }
			}
		}
		if err := dec.Decode(&span); err != nil {
			t.Fatal(err)
		}
		if span.Name != "/v1/items/:id" {
			continue
		}
		for _, a := range span.Attributes {
			if a.Key == "http.user_agent" || a.Key == "http.client_ip" {
				t.Errorf("unexpected private span attribute %s", a.Key)
			}
			if a.Key == "http.target" {
				found = true
				if a.Value.Value != "/v1/items/7" {
					t.Errorf("http.target=%v", a.Value.Value)
				}
			}
		}
	}
	if !found {
		t.Fatal("missing sampled HTTP target")
	}
}
func assertNoAccessSecrets(t *testing.T, text string) {
	t.Helper()
	for _, marker := range []string{"OBS_SEARCH", "OBS_PATH_SECRET", "OBS_PASSWORD", "OBS_TOKEN", "OBS_COOKIE", "OBS_RESPONSE_SECRET", "OBS_BAGGAGE", "OBS_STATE", "OBS_PRIVATE_UA", "203.0.113.77", "192.0.2.24"} {
		if strings.Contains(text, marker) {
			t.Errorf("leaked %s", marker)
		}
	}
}
func TestStreamRequiresResponseControllerSupport(t *testing.T) {
	err := Stream(context.Background(), httptest.NewRecorder(), io.NopCloser(strings.NewReader("payload")), time.Now().Add(time.Second))
	if !errors.Is(err, http.ErrNotSupported) {
		t.Fatalf("got %v", err)
	}
}
