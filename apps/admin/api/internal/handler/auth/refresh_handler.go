// Code scaffolded by goctl. Safe to edit.
package auth

import (
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/identity"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"net/http"
	"time"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/authhttp"
	authlogic "github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/auth"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func RefreshHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.RefreshRequest
		if err := httpx.ParseHeaders(r, &req); err != nil || !identity.ValidRefreshKey(req.RetryKey) {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}
		r = r.WithContext(identity.WithRefreshKey(r.Context(), req.RetryKey))
		raw, err := authhttp.ReadCookie(r)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		l := authlogic.NewRefreshLogic(r.Context(), svcCtx)
		resp, err := l.Refresh(raw)
		if err != nil {
			if status.Code(err) == codes.Unauthenticated {
				authhttp.SetCookie(w, "", time.Time{}, svcCtx.Config.CookieSecure)
			}
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		authhttp.SetCookie(w, resp.RefreshToken, resp.RefreshExpiresAt, svcCtx.Config.CookieSecure)
		httpx.OkJsonCtx(r.Context(), w, resp.Response)
	}
}
