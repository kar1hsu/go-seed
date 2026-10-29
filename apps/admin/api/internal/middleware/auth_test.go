package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kar1hsu/go-seed/pkg/xhttp"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestOriginGuard(t *testing.T) {
	httpx.SetErrorHandlerCtx(xhttp.ErrorHandler)
	guard, err := NewBrowserOriginMiddleware("http://127.0.0.1:8088")
	if err != nil {
		t.Fatal(err)
	}
	handler := guard.Handle(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusNoContent) })
	for _, tt := range []struct {
		name, origin, csrf, site string
		status                   int
	}{
		{"same origin", "http://127.0.0.1:8088", "1", "same-origin", 204},
		{"CLI", "", "1", "", 204},
		{"foreign origin", "https://example.com", "1", "", 403},
		{"null origin", "null", "1", "", 403},
		{"simple request", "http://127.0.0.1:8088", "", "", 403},
		{"cross site", "", "1", "cross-site", 403},
	} {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/admin-api/v1/auth/refresh", nil)
			req.Header.Set("Origin", tt.origin)
			req.Header.Set("X-GoSeed-CSRF", tt.csrf)
			req.Header.Set("Sec-Fetch-Site", tt.site)
			recorder := httptest.NewRecorder()
			handler(recorder, req)
			if recorder.Code != tt.status {
				t.Fatalf("status %d", recorder.Code)
			}
			if recorder.Header().Get("Cache-Control") != "no-store" {
				t.Fatal("missing cache control")
			}
		})
	}
}
