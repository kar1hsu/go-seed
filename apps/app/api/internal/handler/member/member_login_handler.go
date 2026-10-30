// Code scaffolded by goctl. Safe to edit.
package member

import (
	"net/http"

	"github.com/kar1hsu/go-seed/apps/app/api/internal/authhttp"
	authlogic "github.com/kar1hsu/go-seed/apps/app/api/internal/logic/member"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"github.com/zeromicro/go-zero/rest/httpx"
	"google.golang.org/grpc/codes"
)

func MemberLoginHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MemberLoginRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}
		l := authlogic.NewMemberLoginLogic(r.Context(), svcCtx)
		resp, err := l.MemberLogin(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}
		authhttp.SetCookie(w, resp.RefreshToken, resp.RefreshExpiresAt, svcCtx.Config.CookieSecure)
		httpx.OkJsonCtx(r.Context(), w, resp.Response)
	}
}
