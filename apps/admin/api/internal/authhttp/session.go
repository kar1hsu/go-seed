package authhttp

import (
	"context"
	"net/http"
	"time"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/kar1hsu/go-seed/services/system/rpc/pb"
	"github.com/zeromicro/go-zero/core/trace"
	"google.golang.org/grpc/codes"
)

const CookieName = "goseed_refresh"

type Session struct {
	Response         *types.AuthResponse
	RefreshToken     string
	RefreshExpiresAt time.Time
}

func Profile(a *pb.AdminProfile) types.AdminProfile {
	return types.AdminProfile{Id: a.Id, Username: a.Username, DisplayName: a.DisplayName}
}
func Response(ctx context.Context, p *pb.AdminAuthResponse) (*Session, error) {
	expires, err := time.Parse(time.RFC3339, p.RefreshExpiresAt)
	if err != nil || p.Admin == nil {
		return nil, xerr.New(codes.Internal)
	}
	return &Session{Response: &types.AuthResponse{Code: 0, Message: "ok", Data: types.AuthData{AccessToken: p.AccessToken, ExpiresAt: p.ExpiresAt, Admin: Profile(p.Admin)}, TraceId: trace.TraceIDFromContext(ctx)}, RefreshToken: p.RefreshToken, RefreshExpiresAt: expires}, nil
}
func ReadCookie(r *http.Request) (string, error) {
	var value string
	count := 0
	for _, cookie := range r.Cookies() {
		if cookie.Name == CookieName {
			value = cookie.Value
			count++
		}
	}
	if count > 1 {
		return "", xerr.New(codes.Unauthenticated)
	}
	return value, nil
}
func SetCookie(w http.ResponseWriter, value string, expires time.Time, secure bool) {
	maxAge := int(time.Until(expires).Seconds())
	if value == "" {
		maxAge = -1
		expires = time.Unix(1, 0)
	}
	http.SetCookie(w, &http.Cookie{Name: CookieName, Value: value, Path: "/admin-api/v1/auth", Expires: expires, MaxAge: maxAge, HttpOnly: true, Secure: secure, SameSite: http.SameSiteStrictMode})
	w.Header().Set("Cache-Control", "no-store")
}
