// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.2

package member

import (
	"github.com/kar1hsu/go-seed/pkg/xerr"
	"google.golang.org/grpc/codes"
	"net/http"

	"github.com/kar1hsu/go-seed/apps/admin/api/internal/logic/member"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/svc"
	"github.com/kar1hsu/go-seed/apps/admin/api/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Enable or disable a member; disabling revokes all sessions
func SetMemberUserStatusHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.SetMemberUserStatusRequest
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, xerr.New(codes.InvalidArgument))
			return
		}

		l := member.NewSetMemberUserStatusLogic(r.Context(), svcCtx)
		resp, err := l.SetMemberUserStatus(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
