package xhttp

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// WithAccessLog installs the access router and shields go-zero's HTTP tracer
// from client identifiers. Application middleware still sees the original
// values, notably the forwarding chain used by trusted-proxy validation.
func WithAccessLog(service string) rest.RunOption {
	return func(s *rest.Server) {
		rest.WithRouter(privateTraceRouter{NewAccessRouter(service)})(s)
		s.Use(restoreClientHeaders)
	}
}

type privateTraceRouter struct{ httpx.Router }
type clientHeadersKey struct{}
type clientHeaders struct{ forwardedFor, userAgent []string }

func (p privateTraceRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h := clientHeaders{r.Header.Values("X-Forwarded-For"), r.Header.Values("User-Agent")}
	if h.forwardedFor != nil || h.userAgent != nil {
		r = r.WithContext(context.WithValue(r.Context(), clientHeadersKey{}, h))
		r.Header = r.Header.Clone()
		r.Header.Del("X-Forwarded-For")
		r.Header.Del("User-Agent")
	}
	p.Router.ServeHTTP(w, r)
}

func restoreClientHeaders(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if h, ok := r.Context().Value(clientHeadersKey{}).(clientHeaders); ok {
			// A private copy keeps even late or concurrent timeout logging sanitized.
			r = r.WithContext(r.Context())
			r.Header = r.Header.Clone()
			if h.forwardedFor != nil {
				r.Header["X-Forwarded-For"] = h.forwardedFor
			}
			if h.userAgent != nil {
				r.Header["User-Agent"] = h.userAgent
			}
		}
		next(w, r)
	}
}
