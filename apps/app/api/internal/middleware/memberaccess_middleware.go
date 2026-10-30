// Code scaffolded by goctl. Safe to edit.
package middleware

import (
	"net/http"

	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
)

type MemberAccessMiddleware struct{ tokens *identity.Verifier }

func NewMemberAccessMiddleware(tokens *identity.Verifier) *MemberAccessMiddleware {
	return &MemberAccessMiddleware{tokens: tokens}
}
func (m *MemberAccessMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		raw := identity.Bearer(r.Header.Get("Authorization"))
		if _, err := m.tokens.Parse(raw); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.Unauthenticated))
			return
		}
		next(w, r.WithContext(identity.WithToken(r.Context(), raw)))
	}
}
