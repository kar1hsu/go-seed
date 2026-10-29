// Code scaffolded by goctl. Safe to edit.
package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
)

type BrowserOriginMiddleware struct{ origins map[string]bool }

func NewBrowserOriginMiddleware(origins string) (*BrowserOriginMiddleware, error) {
	allowed := make(map[string]bool)
	for _, origin := range strings.Split(origins, ",") {
		origin = strings.TrimSpace(origin)
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || u.User != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Path != "" || u.RawQuery != "" || u.Fragment != "" {
			return nil, fmt.Errorf("invalid allowed origin")
		}
		allowed[origin] = true
	}
	return &BrowserOriginMiddleware{origins: allowed}, nil
}
func (m *BrowserOriginMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		origin := r.Header.Get("Origin")
		// A non-simple header is required even without Origin. No cross-origin CORS is enabled.
		if r.Header.Get("X-GoSeed-CSRF") != "1" || r.Header.Get("Sec-Fetch-Site") == "cross-site" || (origin != "" && !m.origins[origin]) {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.PermissionDenied))
			return
		}
		next(w, r)
	}
}
