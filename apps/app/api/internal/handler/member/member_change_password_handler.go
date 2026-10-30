// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package member

import (
	"github.com/kar1hsu/go-seed/apps/app/api/internal/authhttp"
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"google.golang.org/grpc/codes"
	"net/http"
	"time"

	"github.com/kar1hsu/go-seed/apps/app/api/internal/logic/member"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/app/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Change password and revoke all member sessions
func MemberChangePasswordHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.MemberPasswordRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}

		l := member.NewMemberChangePasswordLogic(r.Context(), svcCtx)
		resp, err := l.MemberChangePassword(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			authhttp.SetCookie(w, "", time.Time{}, svcCtx.Config.CookieSecure)
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
