// Package clientip carries the address established by the HTTP trust boundary.
package clientip

import (
	"context"
	"net"
	"net/http"
	"net/netip"
	"strings"
)

type key struct{}

func With(ctx context.Context, value string) context.Context {
	return context.WithValue(ctx, key{}, value)
}
func From(ctx context.Context) string { value, _ := ctx.Value(key{}).(string); return value }
func Valid(value string) bool {
	ip, err := netip.ParseAddr(value)
	return err == nil && ip.Zone() == "" && ip.Unmap().String() == value
}

// Middleware ignores forwarding headers unless the immediate peer is explicitly trusted.
// It walks X-Forwarded-For from the trusted edge backwards, stopping at the first untrusted hop.
func Middleware(cidrs string) (func(http.HandlerFunc) http.HandlerFunc, error) {
	var trusted []netip.Prefix
	for _, raw := range strings.Split(cidrs, ",") {
		if strings.TrimSpace(raw) == "" {
			continue
		}
		p, err := netip.ParsePrefix(strings.TrimSpace(raw))
		if err != nil {
			return nil, err
		}
		trusted = append(trusted, p)
	}
	contains := func(ip netip.Addr) bool {
		for _, p := range trusted {
			if p.Contains(ip) {
				return true
			}
		}
		return false
	}
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			host, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				host = r.RemoteAddr
			}
			peer, err := netip.ParseAddr(host)
			if err != nil {
				http.Error(w, "invalid peer", http.StatusBadRequest)
				return
			}
			peer = peer.Unmap()
			ip := peer
			if contains(peer) {
				hops := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
				for i := len(hops) - 1; i >= 0 && contains(ip); i-- {
					hop, e := netip.ParseAddr(strings.TrimSpace(hops[i]))
					if e != nil || hop.Zone() != "" {
						ip = peer
						break
					}
					ip = hop.Unmap()
				}
			}
			next(w, r.WithContext(With(r.Context(), ip.String())))
		}
	}, nil
}
